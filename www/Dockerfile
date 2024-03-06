FROM node:20-alpine as stage1

ARG NEWSLETTER_SUBMIT_URL
ARG GOOGLE_SITE_VERIFICATION
ARG SITE

ENV PUBLIC_GOOGLE_SITE_VERIFICATION ${GOOGLE_SITE_VERIFICATION}
ENV SITE ${SITE}
ENV NEWSLETTER_SUBMIT_URL ${NEWSLETTER_SUBMIT_URL}

WORKDIR /app

RUN corepack enable

COPY ./package.json ./pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --prefer-offline

COPY ./astro.config.mjs ./
COPY ./public ./public
COPY ./src ./src

RUN pnpm build

FROM nginx:1.25.1-alpine as final

COPY ./nginx /etc/nginx
COPY --from=stage1 /app/dist /srv/
