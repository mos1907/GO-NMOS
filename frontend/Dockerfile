FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
# We rely on npm install inside the container because the lock file may not always be fully in sync with package.json.
RUN npm install
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
