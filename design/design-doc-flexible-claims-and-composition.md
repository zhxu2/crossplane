# Flexible Resource Claims and Composition
* Owner: Nic Cope (@negz)
* Reviewers: Crossplane Maintainers
* Status: Draft

## Background

Crossplane uses a [class and claim] model to provision and manage resources in
an external system, such as a cloud provider. _External resources_ in the
provider's API are modelled as _managed resources_ in the Kubernetes API server.
Managed resources are considered the domain of _infrastructure operators_;
they're cluster scoped infrastructure like a `Node` or `PersistentVolume`.
_Application operators_ may claim a managed resource for a particular purpose by
creating a namespaced _resource claim_. Managed resources may be provisioned
explicitly before claim time (static provisioning), or automatically at claim
time (dynamic provisioning). The initial configuration of dynamically
provisioned managed resources is specified by a _resource class_.

A managed resource is a _high-fidelity_ representation of its corresponding
external resource. High-fidelity in this context means two things:

* A managed resource maps to exactly one external resource - one API object.
* A managed resource is as close to a direct translation of its corresponding
  external API object as is possible without violating [API conventions].

These properties make managed resources - Crossplane's lowest level
infrastructure primitive - flexible and self documenting. Managed resources in
and of themselves hold few opinions about _how_ they should be used, and are
easily related back to the APIs they represent. This provides a solid foundation
upon which to build Crossplane's multicloud capability.

_Application operators_ are typically prevented by [RBAC] from creating and
modifying managed resources directly; they are instead expected to dynamically
provision the managed resources they require by submitting a resource claim.
Crossplane provides claim kinds for common, widely supported resource variants
like `MySQLInstance` and `KubernetesCluster`. There is a one-to-one relationship
between claims and the managed resources they bind to; a `KubernetesCluster`
claim binds to exactly one `GKECluster` managed resource. However, a solitary
resource is often not particularly useful without supporting infrastructure, for
example:

* An RDS instance may be inaccessible without a security group.
* An Azure SQL instance may be inaccessible without a virtual network rule.
* A GKE, EKS, or AKS cluster (control plane) may not be able to run pods without
  a node group.

Crossplane stacks frequently model this supporting infrastructure (there is a
`SecurityGroup` managed resource, for example) but it cannot be dynamically
provisioned or bound to a resource claim. Instead a cluster operator must
statically provision any supporting managed resources ahead of time, then author
resource classes that reference them. This can be limiting:

* Often supporting resources must reference the managed resource they support,
  for example an Azure `MySQLServerVirtualNetworkRule` must reference the
  `MySQLServer` it applies to. Dynamically provisioned managed resources such as
  a `MySQLServer` have non-deterministic names, making it impossible to create
  a `MySQLServerVirtualNetworkRule` until the `MySQLServer` it must reference
  has been provisioned.
* When a resource class references a statically provisioned managed resource
  every managed resource that is dynamically provisioned using that class will
  reference that specific managed resource. For example if a `GKEClusterClass`
  references a `Subnetwork` then every `GKECluster` dynamically provisioned
  using said class will attempt to share said `Subnetwork`, despite it often
  being desirable to create a unique `Subnetwork` for each dynamically
  provisioned `GKECluster`.

The one-to-one relationship between resource claims and resource classes thus
weakens portability, separation of concerns, and support for [GitOps]. An
infrastructure operator can publish a resource class representing a single
managed resource that an application operator may dynamically provision, but in
the likely event that managed resource requires supporting managed resources to
function usefully the application operator must ask an infrastructure operator
to provision them.

Furthermore, defining a core set of portable resource claims has begun to limit
Crossplane. Resource claims are subject to the [lowest common denominator]
problem of multi-cloud; when a claim may provide configuration inputs that may
be used to match or provision many kinds of managed resource it may support only
the settings that apply to _all_ compatible managed resources. This is in part
why Crossplane defines relatively few resource claim kinds. Meanwhile, it's
possible that an infrastructure operator deploys Crossplane to provide an
opinionated abstraction for the application operators in their organisation and
that said organisation only uses AWS. If this organisation values Crossplane's
separation of concerns but does not need its portability there is no reason that
its application operators should be limited to resource claims that are
constrained by supporting all possible providers.

## Goals

The goals of this proposal are to define in configuration (not code) how
resource classes relate to resource claims, and to enable infrastructure
operators to publish resource classes that compose multiple managed resources.

It is important that the design put forward:

* Allow infrastructure operators to define and discover, through configuration,
  which kinds of resource claim may bind to which kinds of managed resource, and
  how those the spec of those resource claims should be used to dynamically
  provision managed resources.
* Support arbitrary resource claims, defined through configuration (not code) by
  infrastructure stack authors or infrastructure operators.
* Allow a resource claim to dynamically provision and bind to more than one
  managed resource, for example a `GKECluster` and a `NodePool`.
* Limit unnecessary API churn, especially to `v1beta1` managed resource APIs
  which must maintain backward compatibility.
* Maintain the high-fidelity, composable nature of managed resources.
* Maintain support for both static and dynamic provisioning.
* Follow [declarative configuration best practices] and Kubernetes [API
  conventions].
* Empower the separation of concerns between infrastructure and application
  owners.

_How_ users will define resource claim kinds is out of scope for this design.
Defining a new Kubernetes resource kind typically involves authoring a [Custom
Resource Definition] (CRD). A method to simplify and constrain this process is
desirable, and may be explored in future.

While the design put forward by this document must allow the composition of
managed resources, it's important to note the aim is not to enable _arbitrary
composition_. Rather, the goal is to remove the limitations put forward in the
[background] of this document by allowing multiple _primitive_ managed resources
to be composed into a _composite_ managed resource. That is to say, a composite
managed resource would conceptually be _one resource_, like "a Kubernetes
cluster" or "an SQL instance". The design must allow an arbitrary selection of
primitive managed resources to be composed into a composite resource, but is not
intended to address use cases such as composing "all east coast production
infrastructure" or "all the infrastructure needs of an application". Composite
managed resources will be useful binding targets for resource claims, which
represent a specific infrastructure need of an application, such as "a MySQL
instance".

## Proposal

This document proposes Crossplane introduce three new kinds:

* A `ResourceClass` configures how an arbitrary kind of resource claim can be
  used to dynamically provision a single primitive managed resource.
* A `CompositeResourceClass` configures how an arbitrary kind of resource claim
  can be used to dynamically provision a composite managed resource, consisting
  of multiple primitive managed resources.
* A `CompositeResource` represents multiple primitive managed resources that
  have been conceptually composed into a single, composite resource.

Note that this document frequently uses the terms _primitive_ and _composite_ to
describe managed resources. These terms are used in the sense of [primitive] and
[composite] data types. Primitive managed resources are those defined, in code,
by a Crossplane infrastructure stack. A composite managed resource can be
thought is conceptually a single managed resource, but one that is composed from
multiple primitive managed resources.

### Example: Google Cloud SQL Instance

In this example a `ResourceClass` configures how to dynamically provision a
`CloudSQLInstance` primitive resource given an infrastructure owner defined
`MySQLInstance` resource claim. Note that said `MySQLInstance` claim is in the
`database.example.org` API group (not `database.crossplane.io`). It is distinct
from the contemporary `MySQLInstance` claim and includes extra spec fields.

```yaml
---
apiVersion: database.example.org/v1alpha1
kind: MySQLInstance
metadata:
  namespace: default
  name: sql
  annotations:
    # The external name annotation is copied to all managed resources that are
    # dynamically provisioned by this claim.
    crossplane.io/external-name: simple-database
spec:
  # The contemporary MySQLInstance resource claim supports engineVersion, but
  # not storageGB and region.
  engineVersion: "5.7"
  storageGB: 10
  region: us-west
  # Resource claim custom resources must support Crossplane's claim binding
  # metadata; they must have a writeConnectionSecretToRef, classRef,
  # resourceRef, bindingPhase, and status conditions.
  writeConnectionSecretToRef:
    name: sql
  # All resource claims continue to support default resource classes, and
  # resource class selection via the claims' classSelector, but we use an
  # explicit class reference for this example.
  classRef:
    apiVersion: infrastructure.crossplane.io/v1alpha1
    kind: ResourceClass
    name: simple-mysql-server
---
apiVersion: infrastructure.crossplane.io/v1alpha1
kind: ResourceClass
metadata:
  name: simple-mysql-server
spec:
  # Each ResourceClass defines how a managed resource is dynamically provisioned
  # for exactly one kind of resource claim; there is a one-to-many
  # claim-kind-to-resource-class-instance relationship.
  claim:
    apiVersion: database.example.org/v1alpha1
    kind: MySQLInstance
  # The primitive managed resource that will be dynamically provisioned to
  # satisfy MySQLInstance claims that use this class, specified as an object
  # containing a 'base' and an optional 'patch'. The base must be a valid
  # managed resource (i.e. a defined custom resource that supports Crossplane's
  # claim binding metadata). The patch specifies how the claim's fields may be
  # used to extend or override the base's fields.
  resource:
    base:
      apiVersion: database.gcp.crossplane.io/v1beta1
      kind: CloudSQLInstance
      # Object metadata is optional. Labels and annotations will be respected
      # when this class is used to instantiate a CloudSQLInstance, but any
      # other fields (including name) will be ignored if specified. Dynamically
      # provisioned managed resources will continue to use generateName, for a
      # name like claimnamespace-claimname-r2l9d. This pattern exists today in
      # Kubernetes, e.g. in a podTemplateSpec.
      metadata:
        labels:
          region: us-west2
      spec:
        forProvider:
          # This embedded CloudSQLInstance is validated against the actual
          # CloudSQLInstance CRD, so any fields required by a CloudSQLInstance
          # are also required here, even if they will always be overwritten by
          # a patch from the resource claim.
          region: us-west2
          databaseVersion: MYSQL_5_6
          settings:
            tier: db-n1-standard-1
            dataDiskType: PD_SSD
            dataDiskSizeGb: 10
            ipConfiguration:
              ipv4Enabled: true
        writeConnectionSecretToRef:
          namespace: crossplane-system
        providerRef:
          name: example
        reclaimPolicy: Delete
    # Each patch entry relates a claim field to a resource field, by fieldPath.
    patches:
    # Overwrite the resource's spec.forProvider.dataDiskSizeGb field with the
    # value of the claim's spec.storageGb field.
    - fromClaimFieldPath: "spec.storageGB"
      toResourceFieldPath: "spec.forProvider.dataDiskSizeGb"
    # Overwrite the resource's spec.forProvider.databaseVersion field with a
    # value corresponding to the claim's spec.engineVersion field, per the
    # supplied map transform function.
    - fromClaimFieldPath: "spec.engineVersion"
      toResourceFieldPath: "spec.forProvider.databaseVersion"
      # Patch entries may include an optional array of transform functions. This
      # allows concepts (like engineVersion below) to be represented in a
      # portable format at the claim level, and translated to the differing
      # formats of each kind managed resource the claim may correspond to. For
      # example "5.6" might map to "MYSQL_5_6" in GCP and "5.6.1" in AWS.
      transforms:
      - type: map
        map:
          "5.6": MYSQL_5_6
          "5.7": MYSQL_5_7
```

Submitting the above `MySQLInstance` resource claim will result in the dynamic
provisioning of and binding to a `CloudSQLInstance`, as configured by the above
`ResourceClass`. This is effectively the same functionality Crossplane provides
today, except that the dynamic provisioning logic has been moved from code into
configuration.

### Example: Azure MySQL Server with Virtual Network Rule

The below example uses the infrastructure operator defined `MySQLInstance` claim
from the [above example] along with a `CompositeResourceClass`. The use of a
`CompositeResourceClass` rather than a `ResourceClass` signals that multiple
primitive managed resources must be provisioned to satisfy the resource claim.

```yaml
---
apiVersion: database.example.org/v1alpha1
kind: MySQLInstance
metadata:
  namespace: default
  name: sql
  annotations:
    crossplane.io/external-name: very-private-database
spec:
  engineVersion: "5.7"
  storageGB: 10
  region: us-west
  writeConnectionSecretToRef:
    name: sql
  classRef:
    apiVersion: infrastructure.crossplane.io/v1alpha1
    kind: CompositeResourceClass
    name: private-mysql-server
---
apiVersion: infrastructure.crossplane.io/v1alpha1
kind: CompositeResourceClass
metadata:
  name: private-mysql-server
spec:
  # Each CompositeResourceClass defines how resources are dynamically
  # provisioned for exactly one kind of resource claim; there is a one-to-many
  # claim-kind-to-composite-resource-class-instance relationship.
  claim:
    apiVersion: database.example.org/v1alpha1
    kind: MySQLInstance
  # The array of primitive managed resources that will be dynamically
  # provisioned to satisfy MySQLInstance claims that use this class. Like the
  # above ResourceClass, each entry is an object containing a 'base' and an
  # optional 'patch'.
  resources:
  - base:
      apiVersion: azure.crossplane.io/v1alpha3
      kind: ResourceGroup
      metadata:
      spec:
        location: West US
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: "spec.region"
      toResourceFieldPath: "spec.forProvider.location"
      transforms:
      - type: map
        map:
          us-west: "West US"
          us-east: "East US"
  - base:
      apiVersion: database.azure.crossplane.io/v1beta1
      kind: MySQLServer
      spec:
        forProvider:
          administratorLogin: myadmin
          # A resource selector allows a resource dynamically provisioned using
          # a particular composite resource class to reference other resources
          # provisioned by the same class. In this case MySQLServer resources
          # provisioned using this class will select the ResourceGroup
          # provisioned by the same class and claim.
          resourceGroupNameSelector:
            matchComposite: true
          location: West US
          sslEnforcement: Disabled
          version: "5.6"
          sku:
            tier: Basic
            capacity: 1
            family: Gen5
          storageProfile:
            storageMB: 20480
        writeConnectionSecretToRef:
          namespace: crossplane-system
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: ".metadata.uid"
      toResourceFieldPath: "spec.writeConnectionSecretToRef.name"
    - fromClaimFieldPath: "spec.engineVersion"
      toResourceFieldPath: "spec.forProvider.version"
    - fromClaimFieldPath: "spec.storageGB"
      toResourceFieldPath: "spec.forProvider.storageMB"
      transforms:
      - type: math
        math:
          multiply: 1024
    - fromClaimFieldPath: "spec.region"
      toResourceFieldPath: "spec.forProvider.location"
      transforms:
      - type: map
        map:
          us-west: "West US"
          us-east: "East US"
  - base:
      apiVersion: database.azure.crossplane.io/v1alpha3
      kind: MySQLServerVirtualNetworkRule
      spec:
        name: my-cool-vnet-rule
        serverNameSelector:
          matchComposite: true
        resourceGroupNameSelector:
          matchComposite: true
        properties:
          virtualNetworkSubnetIdRef:
            name: sample-subnet
        reclaimPolicy: Delete
        providerRef:
          name: azure-provider
```

The below `CompositeResource` is produced by the above `MySQLInstance` and
`CompositeResourceClass`. A `CompositeResource` aggregates distinct primitive
managed resources into a conceptual composite managed resource. This provides a
target that may be statically provisioned for resource claims to later claim and
bind. It also serves to represent the composition of any primitive managed
resources that remain (due to the `Release` reclaim policy) after a resource
claim is deleted.

```yaml
apiVersion: core.crossplane.io/v1alpha1
kind: CompositeResource
metadata:
  # The composite resource is cluster scoped with a name derived from the
  # claim's namespace and name, just like a primitive managed resource.
  name: default-sql-gdm0w
  annotations:
    # Composite resources use annotations to declare which resource claim kinds
    # may claim and bind them. Dynamically provisioned composite resources such
    # as this one infer which resource claim kind may bind them from the
    # composite resource class used to provision them. They can be bound by only
    # one kind of claim because their composite resource class can be used to
    # dynamically provision managed resources for only one kind of claim.
    # Statically provisioned CompositeResources may choose to be bindable by
    # more than one kind of resource claim.
    mysqlinstance.database.example.org/v1alpha1: bindable
spec:
  # This composite resource matches any primitive managed resource with the
  # below labels. Dynamically provisioned composite resources match on their
  # claim's UID to ensure they never accidentally match the wrong resources.
  resourceSelector:
    matchLabels:
      resourceclaim.crossplane.io/namespace: default
      resourceclaim.crossplane.io/name: sql
      resourceclaim.crossplane.io/uid: eabce854-0cd7-11ea-8d71-362b9e155667
  claimRef:
    apiVersion: database.example.org/v1alpha1
    kind: MySQLInstance
    namespace: default
    name: sql
  classRef:
    apiVersion: infrastructure.crossplane.io/v1alpha1
    kind: CompositeResourceClass
    name: private-mysql-server
  writeConnectionSecretsToRef:
    namespace: crossplane-system
    name: eabce854-0cd7-11ea-8d71-362b9e155667
status:
  # Any primitive managed resource matching this composite resource's labels
  # will bind to it and be included in this array of resources. This ensures a
  # one-to-many composite-to-primitive relationship that is discoverable from
  # both the composite and primitive end, unlike Kubernetes owner references.
  resources:
  - apiVersion: azure.crossplane.io/v1alpha3
    kind: ResourceGroup
    name: default-sql-ab4k8
    bindingPhase: Bound
  - apiVersion: database.azure.crossplane.io/v1beta1
    kind: SQLServer
    name: default-sql-d82nd
    bindingPhase: Bound
  - apiVersion: database.azure.crossplane.io/v1alpha3
    kind: MySQLServerVirtualNetworkRule
    name: default-sql-9dm3v
    bindingPhase: Bound
  bindingPhase: Bound
```

Note that the `CompositeResource` binds to the `MySQLInstance` claim, while the
dynamically provisioned primitive managed resources bind in turn to the
`CompositeResource`, thus creating a `resource claim -> composite resource ->
primitive resources` binding relationship. The below `SQLServer` binds to a
`CompositeResource` (per its `compositeRef` and `bindingPhase`) rather than
to a resource claim:

```yaml
apiVersion: database.azure.crossplane.io/v1beta1
kind: SQLServer
metadata:
  name: default-sql-3d093
  labels:
    # These labels are added to all primitive managed resources that are
    # dynamically provisioned as part of a composite resource, allowing the
    # composite resource to match them.
    resourceclaim.crossplane.io/namespace: default
    resourceclaim.crossplane.io/name: sql
    resourceclaim.crossplane.io/uid: eabce854-0cd7-11ea-8d71-362b9e155667
spec:
  forProvider:
    administratorLogin: myadmin
    resourceGroupNameSelector:
      matchComposite: true
    location: West US
    sslEnforcement: Disabled
    version: "5.6"
    sku:
      tier: Basic
      capacity: 1
      family: Gen5
    storageProfile:
      storageMB: 20480
  writeConnectionSecretsToRef:
    namespace: crossplane-system
    name: default-sql-3d093
  providerRef:
    name: example
  reclaimPolicy: Delete
  # Primitive managed resources that are part of a composite resource use a
  # compositeRef, instead of a claimRef. This signals that the resource is not
  # available for binding to a managed resource, and must instead be bound via
  # its composite resource. It also allows cross-resource references to be set
  # by matching composite references.
  compositeRef:
    apiVersion: core.crossplane.io/v1alpha1
    kind: CompositeResource
    name: default-sql-gdm0w
```

### Example: GKE Cluster with Node Pools

The below example uses a new infrastructure operator defined resource claim
(`KubernetesCluster`) along with a `CompositeResourceClass` to produce a
`CompositeResource` consisting of a GKE cluster control plane with three node
pools.

```yaml
---
apiVersion: compute.example.org/v1alpha1
kind: KubernetesCluster
metadata:
  namespace: default
  name: cluster
spec:
  version: 1.16
  writeConnectionSecretToRef:
    name: cluster
  classRef:
    apiVersion: infrastructure.crossplane.io/v1alpha1
    kind: CompositeResourceClass
    name: gke-cluster
---
apiVersion: infrastructure.crossplane.io/v1alpha1
kind: CompositeResourceClass
metadata:
  name: gke-cluster
spec:
  claim:
    apiVersion: compute.example.org/v1alpha1
    kind: KubernetesCluster
  resources:
  - base:
      apiVersion: container.gcp.crossplane.io/v1beta1
      kind: GKECluster
      spec:
        forProvider:
          location: us-central1
          initialClusterVersion: 1.16
          masterAuth:
            username: admin
            clientCertificateConfig:
              issueClientCertificate: true
        writeConnectionSecretsToRef:
          namespace: crossplane-system
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: ".metadata.uid"
      toResourceFieldPath: "spec.writeConnectionSecretToRef.name"
    - fromClaimFieldPath: "spec.version"
      toResourceFieldPath: "spec.forProvider.version"
  - base:
      apiVersion: container.gcp.crossplane.io/v1alpha3
      kind: NodePool
      metadata:
        annotations:
          # Composite classes that template more than one of a particular
          # primitive managed resource kind must set the external-name-prefix or
          # suffix annotation in order to ensure multiple external resources of
          # the same type are not created with the same name. These annotations
          # will be set on the the provisioned managed resources and parsed by
          # their controllers.
          crossplane.io/external-name-suffix: "-a"
      spec:
        forProvider:
          clusterSelector:
            matchComposite: true
          autoscaling:
            enabled: true
            minNodeCount: 0
            maxNodecount: 100
          config:
            diskType: PD_SSD
            machineType: n1-standard-8
          version: 1.16
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: "spec.version"
      toResourceFieldPath: "spec.forProvider.version"
  # Note that the composite class includes three node pool resources, two of
  # which are identical. This design does not support a claim author requesting
  # a configurable number of primitive managed resources be  provisioned using a
  # particular managed resource template. This is intentional; doing so would
  # both complicate the configuration and blur the separation of concerns by
  # shifting infrastructure concerns (how many node pools a cluster should have)
  # away from the infrastructure owner and onto the application owner.
  - base:
      apiVersion: container.gcp.crossplane.io/v1alpha3
      kind: NodePool
      metadata:
        annotations:
          crossplane.io/external-name-suffix: "-b"
      spec:
        forProvider:
          clusterSelector:
            matchComposite: true
          autoscaling:
            enabled: true
            minNodeCount: 0
            maxNodecount: 100
          config:
            diskType: PD_SSD
            machineType: n1-standard-8
          version: 1.16
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: "spec.version"
      toResourceFieldPath: "spec.forProvider.version"
  - base:
      apiVersion: container.gcp.crossplane.io/v1alpha3
      kind: NodePool
      metadata:
        annotations:
          crossplane.io/external-name-suffix: "-highmem"
      spec:
        forProvider:
          clusterSelector:
            matchComposite: true
          autoscaling:
            enabled: true
            minNodeCount: 0
            maxNodecount: 100
          config:
            diskType: PD_SSD
            machineType: n1-highmem-32
          version: 1.16
        providerRef:
          name: example
        reclaimPolicy: Delete
    patch:
    - fromClaimFieldPath: "spec.version"
      toResourceFieldPath: "spec.forProvider.version"
```

### Transform Functions

TODO(negz): Explain transform functions.

* Simple, schemafied, functions that can transform claim fields.
* Similar to Terraform's small set of functions.
* Purposefully not open ended.
* Written using plain old YAML.

### Composite References

Crossplane allows certain fields of primitive managed resources to be set to a
value inferred from another primitive managed resource using [cross resource
references]. Take for example a resource that must specify the VPC network in
which it should be created, by specifying the `network` field:

```yaml
spec:
  forProvider:
    network: /projects/example/global/networks/desired-vpc-network
```

Cross resource references allow this resource to instead reference the `Network`
managed resource that represents the desired VPC Network:

```yaml
spec:
  forProvider:
    networkRef:
      name: desired-vpc-network
```

The managed resource reconcile logic resolves this reference and populates the
`network` field, resulting in the following configuration:

```yaml
spec:
  forProvider:
    # Network is populated with the value calculated by networkRef.
    network: /projects/example/global/networks/desired-vpc-network
    networkRef:
      name: desired-vpc-network
```

This functionality must be extended in order to support composite resource
classes. Consider a `CompositeResourceClass` that may be used to dynamically
provision the following primitive managed resources:

* `Subnetwork` A
* `GKECluster` A
* `ServiceAccount` A
* `ServiceAccount` B
* `GKENodePool` A
* `GKENodePool` B

The `CompositeResourceClass` author would like to configure the resources such
that:

1. `Subnetwork` A is created in an existing, statically provisioned `Network`.
1. `GKECluster` A is created in `Subnetwork` A.
1. `GKENodePool` A uses `ServiceAccount` A.
1. `GKENodePool` B uses `ServiceAccount` B.
1. Both `GKENodePool` resources join `GKECluster` A.

The author cannot use a contemporary cross resource reference for requirements
two through five. Managed resources are referenced by name, and the names of
dynamically provisioned resources are non-deterministic; they are not known
until they have been provisioned.

This document proposes the introduction of a reference _selector_, which allows
a managed resource to describe the properties of the distinct resource it wishes
to reference, rather than explicitly naming it.

```yaml
spec:
  forProvider:
    networkSelector:
      # Match only managed resources that are part of the same composite, i.e.
      # managed resources that have the same compositeRef as the selecting
      # resource.
      matchComposite: true
      # Match only managed resources with the supplied labels.
      matchLabels:
        example: label
```

The combination of these two fields allows a managed resource to uniquely
identify a distinct managed resource within the same composite. In the previous
example the `GKENodePool` resources need only use `matchComposite` to match the
`GKECluster` they wish to join, because there is only one `GKECluster` for them
to match within their composite resource. They need to use both `matchComposite`
and `matchLabels` to match their desired `ServiceAccount`; the labels
distinguish of the two composed `ServiceAccount` resources are matched.

If a reference field is set, its corresponding selector field is ignored. If the
selector field is unset, it is ignored. If the specified selector matches
multiple managed resources one is chosen at random, though specifying both
`matchComposite` and `matchLabels` can always guarantee that at most one
provisioned managed resource will match the selector.

### External Names

Managed resources have a name and an _external name_. The former identifies the
managed resource in the Kubernetes API, while the latter identifies the resource
in the external system. Managed resources that have reached `v1beta1` allow both
managed resource and resource claim authors to [control the name of their
external resource] using the `crossplane.io/external-name` annotation:

* The name of the external resource is set to the value of the annotation.
* If the annotation is absent it is set to the managed resource's name.
* The annotation is propagated from a resource claim to any managed resource it
  dynamically provisions.

This final point may pose a problem when one resource claim may provision many
managed resources. If a resource claim set the `crossplane.io/external-name`
annotation and referenced a `CompositeResourceClass` that specified two
managed resources of the same kind, both dynamically provisioned resources would
attempt to use the same external name. External names are unique at different
scopes (global, project, region) for different resources so this will not be an
issue all resources, but it may for some.

This document proposes two new external naming annotations be introduced to help
configuration class authors avoid external name conflicts:

* `crossplane.io/external-name-prefix` - A value prepended to the external name
  of a managed resource.
* `crossplane.io/external-name-suffix` - A value appended to the external name
  of a managed resource.

These annotations would be set on the provisioned managed resources and parsed
by their controllers alongside the existing `crossplane-io/external-name`
annotation, resulting in an actual external name of the form
`{external-name-prefix}{external-name}{external-name-suffix}`.

### Connection Secrets

TODO(negz): Determine a solution for connection secrets.

Currently the one-to-one claim-to-resource binding allows us to simply propagate
the resource's connection secret (if any) to the claim. With a one-to-many
claim-to-resource binding we need to handle the case in which there is more than
one secret to propagate back to the claim.

Possible options include:

* Forming one secret from many, per the defunct [aggregate resource design](https://github.com/crossplaneio/crossplane/pull/1094/files#diff-054f386cc6cfafffb8cf96a031552573R362).
* Making the infra operator specify exactly one secret to be propagated to the
  claim.

### Backward Compatibility

TODO(negz): Explain how this design is backward compatible with our current
controllers. I'm _pretty sure_ this design is backward compatible, and could
thus live alongside our existing resource classes, claims, etc. This would allow
us to deprecate the existing classes (and claim controllers) and migrate away
from them cleanly.

## Alternatives Considered

The following alternatives were considered in arriving at the proposal put
forward by this document.

### Opaque Resource Templates

TODO(negz): Explain why we didn't take this direction

Note that this design is intentionally open to hybrid approaches, for example:

* A `CompositeResourceClass` could produce a managed resource that happened
  to be implemented by a template stack, and thus used a template to produce
  several primitive managed resources.
* It may be possible to support optional opaque templates in this design for
  those who wished to trade safety for [more power](https://youtu.be/IelyPhtPimA), e.g:

```yaml
apiVersion: infrastructure.crossplane.io/v1alpha1
kind: CompositeResourceClass
metadata:
  name: gke-cluster
spec:
  claim:
    apiVersion: compute.example.org/v1alpha1
    kind: KubernetesCluster
  resources:
  # This resource is produced using a template, rather than a base and patch. It
  # is assumed to produce the equivalent of a patched base; a valid Crossplane
  # managed resource. If it does not, dynamic provisioning will fail. A template
  # based resource can be mixed in with other base-and-patch style resources.
  - template:
      engine: go-sprig
      source:
        type: inline
        inline: |
          apiVersion: database.azure.crossplane.io/v1beta1
          kind: MySQLServer
          spec:
            forProvider:
              administratorLogin: myadmin
              resourceGroupNameSelector:
                matchComposite: true
              location: West US
              sslEnforcement: Disabled
              version: "{{ spec.version }}"
              sku:
                tier: Basic
                capacity: 1
                family: Gen5
              storageProfile:
                storageMB: 20480
            writeConnectionSecretToRef:
              namespace: crossplane-system
              name: {{ .metadata.uid }}
            providerRef:
              name: example
            reclaimPolicy: Delete
```

### Aggregate Managed Resources

https://github.com/crossplaneio/crossplane/pull/1094

TODO(negz): Explain why we didn't take this direction

### Resource Binding Definition

https://github.com/crossplaneio/crossplane/pull/1118

TODO(negz): Explain why we didn't take this direction

### Complex Managed Resources

One alternative to resource configurations would be to model tightly coupled
external resources as a single managed resource. For example the `MySQLServer`
managed resource might allow the author to configure an array of virtual network
rules in its spec. Such "complex" managed resources are not uncommon in
Crossplane's older controllers; at the time of writing creating an `S3Bucket`
managed resource for example creates both an S3 bucket and an IAM user in the
AWS API.

This design compromises on the "granularity" aspect of high-fidelity managed
resources, as discussed in the [Background] section of this document. While this
may seem innocuous, it has undesirable properties:

* It forces Crossplane's assumptions on how resources will be used; i.e. that a
  Crossplane user would never want to create a GKE node pool without also
  managing its control plane in Crossplane, or would never want to create a
  virtual network rule for an Azure SQL server that Crossplane was unaware of.
* It violates the [principle of least astonishment]. In the `S3Bucket` example
  above it may be surprising for a Crossplane user to be provisioned an IAM user
  when they only requested an S3 bucket.
* The mapping from Crossplane managed resource to underlying API becomes less
  obvious - it's not clear that virtual network rules are actually a distinct
  API rather than an inherent part of a MySQL server. This makes it harder for
  Crossplane users to fall back to the underlying provider's documentation when
  Crossplane's is unclear.
* Each managed resource reconciler becomes more complex, as it must reconcile
  each managed resource with multiple external resources.

Supporting complex managed resources would remove (or less generously, move) the
need to add complexity to the Crossplane API and resource claim logic, but the
cost would be a complex mental model of how Crossplane managed resources relate
to cloud resources, and limitations placed on how Crossplane could be used.

### Cross-claim References

Cross-claim references would allow resource claims to reference other resource
claims, similarly to how managed resources can reference other managed
resources. Rather than a `MySQLInstance` claim binding to an `AggregateResource`
of a `MySQLServer` and a `MySQLVirtualNetworkRule` each managed resource would
be claimed separately - the `MySQLServer` by the `MySQLInstance` and the
`MySQLVirtualNetworkRule` by a new claim: perhaps `FirewallRule`. The
`FirewallRule` claim would state via a cross-claim reference that it applies to
the `MySQLInstance` claim.

This design shifts much of the burden of designing infrastructure from
infrastructure operators to application operators. Instead of an application
operator requesting a `MySQLInstance` and trusting that the infrastructure
operator has published a resource class that will ensure it is securely
configured, they must also request a `FirewallRule` be applied to that
`MySQLInstance`. Furthermore, the "lowest common denominator" problem affects
resource claims, in that they can only expose configuration fields that
translate to _every_ managed resource that could satisfy the claim. A
`FirewallRule` for example could only expose configuration fields that apply to
`MySQLVirtualNetworkRule`, `SecurityGroup`, and any other firewall-like managed
resource. This means application owners would simultaneously be saddled with the
burden of designing appropriate infrastructure for their needs and a limited
language in which to do so.

[class and claim]: https://static.sched.com/hosted_files/kccncna19/2d/kcconna19-eric-tune.pdf
[API conventions]: https://github.com/kubernetes/community/blob/862de062acf8bbd84f7a655914fa08972498819a/contributors/devel/sig-architecture/api-conventions.md
[RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[GitOps]: https://www.weave.works/technologies/gitops/
[background]: #background
[lowest common denominator]: https://thenewstack.io/avoiding-least-common-denominator-approach-hybrid-clouds/
[Custom Resource Definition]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[_composite_]: https://en.wikipedia.org/wiki/Composite_data_type
[_primitive_]: https://en.wikipedia.org/wiki/Primitive_data_type
[declarative configuration best practices]: https://github.com/kubernetes/community/blob/5d62001/contributors/design-proposals/architecture/declarative-application-management.md#declarative-configuration
[cross resource reference]: one-pager-cross-resource-referencing.md
[above example]: #example-google-cloud-sql-instance