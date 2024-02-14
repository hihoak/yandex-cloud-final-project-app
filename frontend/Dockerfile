FROM node:18.19.0-slim
# Create app directory
WORKDIR /usr/src/app
COPY *.json *.js ./
COPY src src
COPY public public
COPY node_modules node_modules

ENV NODE_OPTIONS=--openssl-legacy-provider
ENTRYPOINT [ "npm", "run", "serve" ]
EXPOSE 8080
