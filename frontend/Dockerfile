ARG TARGETPLATFORM
FROM --platform=$TARGETPLATFORM node:22-alpine AS build
WORKDIR /src

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM --platform=$TARGETPLATFORM nginx:1.27-alpine
COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]