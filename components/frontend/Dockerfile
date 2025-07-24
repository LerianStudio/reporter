# Use an official Node.js runtime as the base image
FROM node:22-alpine

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to the working directory
COPY package*.json ./

# Install dependencies
RUN apk update && apk add bash && \
    npm install && \
    npm cache clean --force

# Copy .env.example to the working directory
COPY .env.example .env.example

# Run the command to set environment variables (if needed)
RUN npm run set-env

# Copy the rest of the application source code
COPY . .

# Build the app
RUN npm run build

# Expose the port the app runs on
EXPOSE 8083

# Command to run the application
ENTRYPOINT ["npm", "start"]
