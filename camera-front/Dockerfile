FROM node:16.18.0-alpine3.15 AS appbuild
WORKDIR /app/camera-front
COPY package*.json .
RUN npm install
COPY . .
RUN npm run build

FROM node:16.18.0-alpine3.15
RUN npm install pm2 -g
WORKDIR /app/camera-front
COPY node/ ./node
RUN chmod +x /app/camera-front/node/start.sh
WORKDIR /app/camera-front/node
RUN npm install
WORKDIR /app/camera-front
COPY --from=appbuild /app/camera-front/dist ./dist
CMD ["/app/camera-front/node/start.sh"]