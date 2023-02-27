#! /usr/bin/env bash

openssl genrsa -out test.3.CA.key 4096
openssl req -new -out test.3.CA.csr -key test.3.CA.key
openssl x509 -req -in test.3.CA.csr -out test.3.CA.crt -extfile openssl.cnf -extensions v3_ca -signkey test.3.CA.key -CAcreateserial -days 3650
openssl genrsa -out test.3.derive.key 4096
openssl req -new -out test.3.derive.csr -key test.3.derive.key
openssl x509 -req -in test.3.derive.csr -out test.3.derive.crt -CA test.3.CA.crt -CAkey test.3.CA.key -CAcreateserial -days 3650 -extfile test.3.ext
