run: dash
	./dash -MAC ac:63:be:56:a4:7f -cmd './hive.py -n "Office a" "Office b"' -MAC 18:74:2e:c5:04:6b -cmd './hue.py -n "Ceiling"'

dash:
	go build .
	sudo setcap cap_net_raw,cap_net_admin=eip dash

