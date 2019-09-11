# core.crossplane.io/v1alpha1 API Reference

Package v1alpha1 contains core API types used by most Crossplane resources.

This API group contains the following resources:


## BindingPhase

BindingPhase represents the current binding phase of a resource or claim. Alias of string.

Appears in:

* [BindingStatus](#BindingStatus)


## BindingStatus

A BindingStatus represents the bindability and binding of a resource.

Appears in:

* [ResourceClaimStatus](#ResourceClaimStatus)* [ResourceStatus](#ResourceStatus)

Name | Type | Description
-----|------|------------
`bindingPhase` | [BindingPhase](#BindingPhase) | Phase represents the binding phase of the resource.



## Condition

A Condition that may apply to a managed resource.

Appears in:

* [ConditionedStatus](#ConditionedStatus)

Name | Type | Description
-----|------|------------
`type` | [ConditionType](#ConditionType) | Type of this condition. At most one of each condition type may apply to a managed resource at any point in time.
`status` | [core/v1.ConditionStatus](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#conditionstatus-v1-core) | Status of this condition; is it currently True, False, or Unknown?
`lastTransitionTime` | [meta/v1.Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#time-v1-meta) | LastTransitionTime is the last time this condition transitioned from one status to another.
`reason` | [ConditionReason](#ConditionReason) | A Reason for this condition&#39;s last transition from one status to another.
`message` | string | A Message containing details about this condition&#39;s last transition from one status to another, if any.



## ConditionReason

A ConditionReason represents the reason a resource is in a condition. Alias of string.

Appears in:

* [Condition](#Condition)


## ConditionType

A ConditionType represents a condition a resource could be in. Alias of string.

Appears in:

* [Condition](#Condition)


## ConditionedStatus

A ConditionedStatus reflects the observed status of a managed resource. Only one condition of each type may exist. Do not manipulate Conditions directly - use the Set method.

Appears in:

* [ResourceClaimStatus](#ResourceClaimStatus)* [ResourceStatus](#ResourceStatus)

Name | Type | Description
-----|------|------------
`conditions` | [[]Condition](#Condition) | Conditions of the managed resource.



## NonPortableClassSpecTemplate

NonPortableClassSpecTemplate contains standard fields that all non-portable classes should include in their spec template. NonPortableClassSpecTemplate should typically be embedded in a non-portable class specific struct.

Name | Type | Description
-----|------|------------
`providerRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 
`reclaimPolicy` | [ReclaimPolicy](#ReclaimPolicy) | A ReclaimPolicy determines what should happen to managed resources when their bound resource claims are deleted.



## PortableClass

PortableClass contains standard fields that all portable classes should include. Class should typically be embedded in a specific portable class.

Name | Type | Description
-----|------|------------
`classRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | NonPortableClassReference is a reference to a non-portable class.



## ReclaimPolicy

A ReclaimPolicy determines what should happen to managed resources when their bound resource claims are deleted. Alias of string.

Appears in:

* [NonPortableClassSpecTemplate](#NonPortableClassSpecTemplate)* [ResourceSpec](#ResourceSpec)


## ResourceClaimSpec

ResourceClaimSpec contains standard fields that all resource claims should include in their spec. Unlike ResourceClaimStatus, ResourceClaimSpec should typically be embedded in a claim specific struct.

Name | Type | Description
-----|------|------------
`writeConnectionSecretToRef` | [core/v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | 
`classRef` | [core/v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | PortableClassReference is a reference to a portable class by name.
`resourceRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 



## ResourceClaimStatus

ResourceClaimStatus represents the status of a resource claim. Claims should typically use this struct as their status.

Name | Type | Description
-----|------|------------

Supports all fields of [ConditionedStatus](#ConditionedStatus).
Supports all fields of [BindingStatus](#BindingStatus).


## ResourceSpec

ResourceSpec contains standard fields that all resources should include in their spec. ResourceSpec should typically be embedded in a resource specific struct.

Name | Type | Description
-----|------|------------
`writeConnectionSecretToRef` | [core/v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#localobjectreference-v1-core) | 
`claimRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 
`classRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | NonPortableClassReference is a reference to a non-portable class.
`providerRef` | [core/v1.ObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#objectreference-v1-core) | 
`reclaimPolicy` | [ReclaimPolicy](#ReclaimPolicy) | A ReclaimPolicy determines what should happen to managed resources when their bound resource claims are deleted.



## ResourceStatus

ResourceStatus contains standard fields that all resources should include in their status. ResourceStatus should typically be embedded in a resource specific status.

Name | Type | Description
-----|------|------------

Supports all fields of [ConditionedStatus](#ConditionedStatus).
Supports all fields of [BindingStatus](#BindingStatus).


Generated with `gen-crd-api-reference-docs` on git commit `b3f29c58`.