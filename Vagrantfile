# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"

  # Configure VM resources
  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
    v.name = "dev-server"
  end

  # Network configuration
  config.vm.network "private_network", ip: "192.168.56.10"

  # Port forwarding
  config.vm.network "forwarded_port", guest: 22, host: 2222

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

    timedatectl set-timezone Europe/Amsterdam

    # Create deployment user (similar to cloud setup)
    useradd -m -s /bin/bash deploy
    mkdir -p /home/deploy/.ssh
    chown -R deploy:deploy /home/deploy/.ssh

    echo "Development environment ready!"
    echo "Access the VM using: vagrant ssh"
  SHELL
end