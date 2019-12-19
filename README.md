# PROMETHIUM

Promethium is a set of services around the Firecracker Micro VM system developed by Amazon utilising the Linux KVM module. 

Firecracker is ued in production by Amazon and provides a reliable way for deploying containerised applications but utilising the additonal security of confined ivrutal machines.

Promethium aims to provide a convenient layer for managing Firecracker VMs whether in production or ona  development environment and contains the following core components:

- Sever & API endpoint for managing VMs in a secure fashion
- Integration with OSv and Capstan - in fact a wrapped and customised version of capstan is available as part of the command line tool. The bundled Capstan is extended and modified by Promethium to support Firecracker.
- A command line tool to manage Promethium instances
- A library for interacting with Promethium, Firecracker and Capstan - exposed via the API

## Dependencies

Promethiums server API and the Firecracker system only run on linux and require KVM. The Command line tools can be used on any host but must connect to a valid Promthium server. Packages contain everything you need but when cusomising and/or building Promethium it is necessary to bootstrap your environment. Please ensure your system has the following installed:
- A KVM enabled Linux distribution
- Docker
- Go 1.12 or above

## Installation
### From Packages
Visit the following link to install Promethium for your linux distribution
### From Source
It is necessary to obtain the source first and build the support tools. There are some bootstrapping scripts included to help with this.
