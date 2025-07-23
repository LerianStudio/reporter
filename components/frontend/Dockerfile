# Use an official Node.js runtime as the base image
FROM node:22-alpine

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to the working directory
COPY components/frontend/package*.json ./

# Configure npm to use GitHub token temporarily, without exposing it in the Dockerfile
RUN apk update && apk add bash && \
    npm install && \
    npm cache clean --force

# Copiar o arquivo .env.example para o diretório de trabalho
COPY components/frontend/.env.example .env.example

# Run the command to set environment variables
RUN npm run set-env

# Copy the rest of the application code
COPY components/frontend/   .

# Build the app
RUN npm run build

# Expose the port the app runs on
EXPOSE 8083

# Command to run the applications
ENTRYPOINT ["npm", "start"]
