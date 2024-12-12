
dev-vm:
	@vagrant up
	@ssh -i .vagrant/machines/default/virtualbox/private_key -p 2222 vagrant@127.0.0.1