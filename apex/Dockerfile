FROM nginx:alpine

COPY ./nginx/ /etc/nginx
COPY ./static/ /srv/alexghr.me/static

VOLUME [ "/etc/nginx/certs" ]
