# stacks.crossplane.io/v1alpha1 API Reference

Package v1alpha1 contains resources relating to Crossplane Stacks.

This API group contains the following resources:

* [Stack](#Stack)
* [StackRequest](#StackRequest)

## Stack

A Stack that has been added to Crossplane.

Name | Type | Description
-----|------|------------
`apiVersion` | string | `stacks.crossplane.io/v1alpha1`
`kind` | string | `Stack`
`metadata` | [meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectmeta-v1-meta) | Kubernetes object metadata.
`spec` | [StackSpec](#StackSpec) | StackSpec specifies the desired state of a Stack.
`status` | [StackStatus](#StackStatus) | StackStatus represents the observed state of a Stack.



## StackRequest

A StackRequest requests a Stack be added to Crossplane.

Name | Type | Description
-----|------|------------
`apiVersion` | string | `stacks.crossplane.io/v1alpha1`
`kind` | string | `StackRequest`
`metadata` | [meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectmeta-v1-meta) | Kubernetes object metadata.
`spec` | [StackRequestSpec](#StackRequestSpec) | StackRequestSpec specifies the desired state of a StackRequest.
`status` | [StackRequestStatus](#StackRequestStatus) | StackRequestStatus represents the observed state of a StackRequest.



## AppMetadataSpec

AppMetadataSpec defines metadata about the stack application

Appears in:

* [StackSpec](#StackSpec)

Name | Type | Description
-----|------|------------
`title` | string | 
`description` | string | 
`version` | string | 
`icons` | [[]IconSpec](#IconSpec) | 
`maintainers` | [[]ContributorSpec](#ContributorSpec) | 
`owners` | [[]ContributorSpec](#ContributorSpec) | 
`company` | string | 
`category` | string | 
`keywords` | []string | 
`website` | string | 
`source` | string | 
`license` | string | 



## ContributorSpec

ContributorSpec defines a contributor for a stack (e.g., maintainer, owner, etc.)

Appears in:

* [AppMetadataSpec](#AppMetadataSpec)

Name | Type | Description
-----|------|------------
`name` | string | 
`email` | string | 



## ControllerDeployment

ControllerDeployment defines a controller for a stack that is managed by a Deployment.

Appears in:

* [ControllerSpec](#ControllerSpec)

Name | Type | Description
-----|------|------------
`name` | string | 
`spec` | [apps/v1.DeploymentSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#deploymentspec-v1-apps) | 



## ControllerJob

ControllerJob defines a controller for a stack that is installed by a Job.

Appears in:

* [ControllerSpec](#ControllerSpec)

Name | Type | Description
-----|------|------------
`name` | string | 
`spec` | [batch/v1.JobSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#jobspec-v1-batch) | 



## ControllerSpec

ControllerSpec defines the controller that implements the logic for a stack, which can come in different flavors. A golang code (controller-runtime) controller with a managing Deployment is all that is supported currently, but more types will come in the future (e.g., templates, functions/hooks, templates, a new DSL, etc.

Appears in:

* [StackSpec](#StackSpec)

Name | Type | Description
-----|------|------------
`deployment` | [ControllerDeployment](#ControllerDeployment) | 
`job` | [ControllerJob](#ControllerJob) | 



## IconSpec

IconSpec defines the icon for a stack

Appears in:

* [AppMetadataSpec](#AppMetadataSpec)

Name | Type | Description
-----|------|------------
`base64Data` | string | 
`mediatype` | string | 



## LinkSpec

LinkSpec defines a useful link about a stack (e.g., homepage, about page, etc.)

Name | Type | Description
-----|------|------------
`description` | string | 
`url` | string | 



## PermissionsSpec

PermissionsSpec defines the permissions that a stack will require to operate.

Appears in:

* [StackSpec](#StackSpec)

Name | Type | Description
-----|------|------------
`rules` | [[]rbac/v1.PolicyRule](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#policyrule-v1-rbac) | 



## StackRequestSpec

StackRequestSpec specifies the desired state of a StackRequest.

Appears in:

* [StackRequest](#StackRequest)

Name | Type | Description
-----|------|------------
`source` | string | Source is the domain name for the stack registry hosting the stack being requested, e.g., registry.crossplane.io
`package` | string | Package is the name of the stack package that is being requested, e.g., myapp. Either Package or CustomResourceDefinition can be specified.
`crd` | string | CustomResourceDefinition is the full name of a CRD that is owned by the stack being requested. This can be a convenient way of installing a stack when the desired CRD is known, but the package name that contains it is not known. Either Package or CustomResourceDefinition can be specified.



## StackRequestStatus

StackRequestStatus represents the observed state of a StackRequest.

Appears in:

* [StackRequest](#StackRequest)

Name | Type | Description
-----|------|------------
`conditionedStatus` | [v1alpha1.ConditionedStatus](../crossplane-runtime/core-crossplane-io-v1alpha1.md#conditionedstatus) | 
`installJob` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 
`stackRecord` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 



## StackSpec

StackSpec specifies the desired state of a Stack.

Appears in:

* [Stack](#Stack)

Name | Type | Description
-----|------|------------
`customresourcedefinitions` | [CRDList](#CRDList) | CRDList is the full list of CRDs that this stack owns and depends on
`controller` | [ControllerSpec](#ControllerSpec) | ControllerSpec defines the controller that implements the logic for a stack, which can come in different flavors. A golang code (controller-runtime) controller with a managing Deployment is all that is supported currently, but more types will come in the future (e.g., templates, functions/hooks, templates, a new DSL, etc.
`permissions` | [PermissionsSpec](#PermissionsSpec) | PermissionsSpec defines the permissions that a stack will require to operate.

Supports all fields of [AppMetadataSpec](#AppMetadataSpec).


## StackStatus

StackStatus represents the observed state of a Stack.

Appears in:

* [Stack](#Stack)

Name | Type | Description
-----|------|------------
`conditionedStatus` | [v1alpha1.ConditionedStatus](../crossplane-runtime/core-crossplane-io-v1alpha1.md#conditionedstatus) | 
`controllerRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 



Generated with `gen-crd-api-reference-docs` on git commit `b3f29c58`.