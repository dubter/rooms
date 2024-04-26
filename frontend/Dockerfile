FROM node:21-alpine

WORKDIR /app

COPY package*.json ./

RUN npm install
RUN npm install -g nodemon
COPY . .

EXPOSE 3000

ENTRYPOINT ["npm", "run", "dev"]
