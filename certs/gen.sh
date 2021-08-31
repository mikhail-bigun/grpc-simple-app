rm *.pem
# Generate CA's private key and seldf-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=RU/ST=Moscow/L=Moscow/O=Bigun/OU=Software Development/CN=*.mikhailbigun.ru/emailAddress=bigun.md@gmail.com"
openssl x509 -in ca-cert.pem -noout -text 

# Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=RU/ST=Moscow/L=Moscow/O=PC Book/OU=Computer/CN=*.pcbook.com/emailAddress=pcbook@gmail.com"


# Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 365 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf 