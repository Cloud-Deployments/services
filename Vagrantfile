# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"

  # Configure VM resources
  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
    v.name = "runner-dev"
  end

  # Network configuration
  config.vm.network "private_network", ip: "192.168.56.10"

  # Port forwarding
  config.vm.network "forwarded_port", guest: 8080, host: 8080  # Coordinator

  # Shared folder for development
  config.vm.synced_folder ".", "/home/vagrant/dev"

  # Provisioning script
  config.vm.provision "shell", inline: <<-SHELL
    #!/bin/bash
    set -e

    # Update system
    apt-get update
    apt-get upgrade -y

    # Install basic tools
    apt-get install -y \
        git \
        curl \
        wget \
        vim \
        jq \
        tmux \
        build-essential

    # Create test SSH keys
    mkdir -p /home/vagrant/.ssh/test-keys
    ssh-keygen -t rsa -b 4096 -f /home/vagrant/.ssh/test-keys/test_key -N ""
    chown -R vagrant:vagrant /home/vagrant/.ssh

    echo "Local VPS is complete!"
    echo "Access the VM using: vagrant ssh"
  SHELL
end