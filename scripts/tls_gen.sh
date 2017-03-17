#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TLS_DIR="contrib/matchbox/assets/tls"
INI_FILE="$PWD/etc/lazy.ini"

source "${DIR}/utils"

[ -d "$TLS_DIR" ] || mkdir -p $TLS_DIR

cd $TLS_DIR
rm *.pem *.csr 2>/dev/null || true

function new_req_cnf() {
    local role=$1 i=0 domain="$(extract_session_key domain_base)"

    cat - <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.101 = kubernetes
DNS.102 = kubernetes.default
DNS.103 = kubernetes.default.svc
DNS.104 = kubernetes.default.svc.cluster.local
IP.1=10.3.0.1
EOF
    for n in $(extract_nodes); do
        i=$((i+1))
        case $role in
            all)
                echo "DNS.$i=${n}.${domain}"
            ;;
            *)
                if extract_node $n | grep "role=$role" >/dev/null 2>&1; then
                    echo "DNS.$i=${n}.${domain}"
                fi
                ;;
        esac

    done
}

function newCA() {
    local name=$1 role=$2 key="${1}-key.pem" csr="${1}.csr" ca="${1}.pem" csr_cnf="${1}-csr.cnf"
    new_req_cnf $2 > $csr_cnf
    openssl genrsa -out $key 2048
    openssl req -new -key $key -out $csr -subj "/CN=k8s-${name}" -config $csr_cnf
    openssl x509 -req -in $csr -CA "ca.pem" -CAkey "ca-key.pem" -CAcreateserial -out $ca -days 365 -extensions v3_req -extfile $csr_cnf

}

##### Generate CA #####
openssl genrsa -out ca-key.pem 2048
openssl req -x509 -new -nodes -key ca-key.pem -days 10000 -out ca.pem -subj "/CN=k8s-ca"

##### Generate API Server CA #####
newCA api-server master

##### Generate Worker CA #####
newCA worker all
