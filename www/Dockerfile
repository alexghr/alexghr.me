FROM node:lts-alpine as stage1

WORKDIR /app

COPY ./package.json ./yarn.lock ./
RUN yarn install --frozen-lockfile

COPY ./site/ ./site
COPY ./scripts/ ./scripts

RUN yarn run build


FROM nginx:alpine as final

COPY ./nginx /etc/nginx
COPY --from=stage1 /app/public /srv/
