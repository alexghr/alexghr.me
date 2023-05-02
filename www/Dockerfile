FROM node:18-alpine as stage1

ARG GOOGLE_SITE_VERIFICATION
ENV PUBLIC_GOOGLE_SITE_VERIFICATION ${GOOGLE_SITE_VERIFICATION}

WORKDIR /app

RUN corepack enable

COPY ./package.json ./pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --prefer-offline

COPY ./astro.config.mjs ./
COPY ./src ./src

RUN pnpm build

FROM nginx:1.23-alpine as final

COPY ./nginx /etc/nginx
COPY --from=stage1 /app/dist /srv/
