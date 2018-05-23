FROM telegraf:alpine
COPY entrypoint.sh entrypoint.sh
CMD entrypoint.sh 
