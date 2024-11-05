# Vulnerability Detection Project

This project is a distributed system designed to detect vulnerabilities in smart contracts using a combination of gRPC, RabbitMQ, Redis, MySQL, and Nginx. It leverages Docker for containerization and Docker Swarm for orchestration.


## Project Overview
 
### Technologies Used
- **Go (Golang)**: Backend logic and gRPC server implementation.
- **Python**: Vulnerability detection model using libraries like PyTorch.
- **Docker & Docker Swarm**: Containerization and orchestration.
- **RabbitMQ**: Message broker for asynchronous task handling.
- **Redis**: In-memory data store for caching.
- **MySQL**: Relational database for data storage.
- **Nginx**: Reverse proxy and load balancing.
- **gRPC**: Communication protocol between client and server.

## Prerequisites

Ensure the following tools are installed on your machine:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Docker Swarm](https://docs.docker.com/engine/swarm/)
- [Git](https://git-scm.com/)

## Project Setup

### 1. Clone the Repository

Clone the project repository from GitHub to your local machine:
```bash | mac
- git clone https://github.com/lzk12345678/Vul_Detection_Golang.git
- cd repository/backend
```

Because of some reasons about the vulnerability detection model of the code I will not give, interested readers can contact the author, but the project needs to use the image file has been uploaded to dockerhub can be directly downloaded and then test is also possible
### 2. Build Docker Images

docker build -t lzksdocker/server_app -f Dockerfile.server .

docker build -t lzksdocker/backend-ginweb -f Dockerfile.ginweb .

docker build -t lzksdocker/backend-client -f Dockerfile.client .

### 3. Pushing Images to Docker Hub

docker push lzksdocker/server_app:latest
docker push lzksdocker/backend-ginweb:latest
docker push lzksdocker/backend-client:latest

docker pull lzksdocker/server_app:latest
docker pull lzksdocker/backend-ginweb:latest
docker pull lzksdocker/backend-client:latest

### 4. Setting Up Docker Swarm
#### Initialize Docker Swarm on your main node:
- docker swarm init


#### To add other nodes to your swarm, you will get a command similar to:
- docker swarm join --token <token> <manager-ip>:2377

Execute this command on your worker nodes to add them to the swarm.

### 5. Deploying the Stack
#### Deploy the entire stack using the Docker Compose file with Docker Swarm:
- docker stack deploy --compose-file docker-compose.yml my_stack

### 6. Monitoring and Managing the Services
#### Check the status of your services using:
```bash
docker service ls
View detailed information about each service's tasks:

bash
docker service ps my_stack_server_app
```

### 7. Scaling the Services
To scale a service, such as the my_stack_server_app, to run multiple replicas, use:
```bash
docker service scale my_stack_server_app=3
```

### 8. Interacting with the System
After deployment, you can test the system using curl commands to make requests to the detection services.

Example Curl Command to Submit Smart Contract Code:
```bash
curl -X POST http://<your-server-ip>:6060/detect -H "Content-Type: application/json" -d '{
  "contractSourcecode": "pragma solidity ^0.4.24;\n\ncontract ReentrancyVulnerable {\n    mapping(address => uint) public balances;\n\n    function deposit() public payable {\n        balances[msg.sender] += msg.value;\n    }\n\n    function withdraw(uint _amount) public {\n        require(balances[msg.sender] >= _amount, \"Insufficient balance\");\n\n        if (msg.sender.call.value(_amount)()) {\n            balances[msg.sender] -= _amount;\n        }\n    }\n}"
}'
Expected response if a vulnerability is detected:
json
{"result":"检测到漏洞|12, 11, 9"}

Submitting Another Contract Example:
bash
curl -X POST http://<your-server-ip>:6060/detect -H "Content-Type: application/json" -d '{
    "contractSourcecode": "pragma solidity ^0.4.24;\n\ncontract EtherollCrowdfund{\n\n    mapping (address => uint) public balanceOf;\n\n    function calcRefund(address _addressToRefund) internal {\n        uint amount = balanceOf[_addressToRefund];\n\n        if (amount > 0) {\n            if (_addressToRefund.call.value(amount)()) {\n                balanceOf[_addressToRefund] = 0;\n            } else {\n                balanceOf[_addressToRefund] = amount;\n            }\n        } \n    }\n}"
}'
Expected response:
json
{"result":"检测到漏洞|7, 13, 10, 9"}
```
### 9. Using Nginx for Load Balancing
Make sure Nginx is correctly set up to distribute incoming traffic among the multiple replicas of the Gin Web Server to ensure balanced loads and optimal performance.

### 10. Stopping the Docker Swarm Cluster
To remove the stack and stop the cluster:
```bash
docker stack rm my_stack
docker swarm leave --force
```
### 11. Troubleshooting and Issues
If a service fails to start, check the logs using:
```bash
docker service logs my_stack_server_app
For persistent issues with nodes or services, inspect the node statuses:
bash
docker node ls
```

## Additional Notes
### Managing Large Files
Since GitHub limits file sizes to 100MB, the larger files (demo and codeBert_model directories) used for the vulnerability detection model are not included in this repository. These should be handled separately and incorporated into the deployment process manually

## Final Remarks
This project leverages the latest technologies in containerization, orchestration, and service management to create a scalable vulnerability detection system. Docker Swarm ensures efficient use of resources, while RabbitMQ and Redis handle communication and caching, making the system highly responsive and efficient.

Feel free to open issues on GitHub if you encounter any problems or have suggestions for improvements.

