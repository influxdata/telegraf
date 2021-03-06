docker build -t local_telegraf -f scripts/alpine.docker .

docker-compose -f plugins/inputs/kibana/test_environment/docker-compose.yml up
