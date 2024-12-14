
run:
	@docker compose down
	@docker compose build --no-cache coordinator-service runner-service
	@docker compose up --watch
dev-vm:
	@vagrant up
	@ssh -i .vagrant/machines/default/virtualbox/private_key -p 2222 vagrant@127.0.0.1