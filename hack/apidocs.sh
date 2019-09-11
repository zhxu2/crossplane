#! /bin/bash

gendoc() {
    local org=$1
    local project=$2
    local dir=$3
    local package=$4
    local version=$5

    local group_name=$(grep groupName ${GOPATH}/src/github.com/${org}/${project}/${dir}/${package}/${version}/doc.go | awk -F= '{ print $2 }')

    # TODO(negz): Commit local changes to this tool.
    GO111MODULE=on go run ${GOPATH}/src/github.com/ahmetb/gen-crd-api-reference-docs/main.go -v=2 \
        -config hack/apidocconfig/${org}/${project}.json \
        -template-dir hack/apidoctemplate \
        -api-dir "github.com/${org}/${project}/${dir}/${package}/${version}" \
        -out-file docs/api/${org}/${project}/${group_name//./-}-${version}.md
}

# TODO(negz): Currently you need to delete vendor to make this pick up the
# package from your GOPATH. Fix that.
gendoc crossplaneio crossplane-runtime apis core v1alpha1

gendoc crossplaneio crossplane apis cache v1alpha1
gendoc crossplaneio crossplane apis compute v1alpha1
gendoc crossplaneio crossplane apis database v1alpha1
gendoc crossplaneio crossplane apis stacks v1alpha1
gendoc crossplaneio crossplane apis storage v1alpha1
gendoc crossplaneio crossplane apis workload v1alpha1

# TODO(negz): Support the base aws/apis/v1alpha2 package
gendoc crossplaneio stack-aws aws/apis cache v1alpha2
gendoc crossplaneio stack-aws aws/apis compute v1alpha2
gendoc crossplaneio stack-aws aws/apis database v1alpha2
gendoc crossplaneio stack-aws aws/apis storage v1alpha2
