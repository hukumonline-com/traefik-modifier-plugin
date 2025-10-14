FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY test-backend.js ./
EXPOSE 3000
CMD ["npm", "start"]