FROM nginx:1.23-alpine

COPY ./nginx/ /etc/nginx
#COPY ./static/ /srv/alexghr.me/static

EXPOSE 80 443 8448

VOLUME [ "/etc/nginx/certs" ]
