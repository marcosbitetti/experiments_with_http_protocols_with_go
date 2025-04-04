default:
	ls

certificates:
	# openssl genpkey -algorithm RSA -out cmd/server.key -pkeyopt rsa_keygen_bits:2048
	# openssl req -new -key cmd/server.key -out cmd/server.csr -subj "/CN=localhost"
	# openssl x509 -req -in cmd/server.csr -signkey cmd/server.key -out cmd/server.crt -days 365 -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1")
	#
	# echo "after this, run:"
	# echo "sudo cp cmd/server.crt /usr/local/share/ca-certificates/"
	# echo "sudo update-ca-certificates"

force-stop:
	kill -9 $(lsof -t -i tcp:8080)

buffer-sizes:
	echo "run it with sodo to run http3"
	sysctl -w net.core.rmem_max=7500000
	sysctl -w net.core.wmem_max=7500000