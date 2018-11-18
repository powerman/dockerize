FROM ubuntu:18.04

RUN apt-get update && apt-get install -y curl nginx

RUN curl -sfL https://github.com/powerman/dockerize/releases/download/v0.6.2/dockerize-`uname -s`-`uname -m` | install /dev/stdin /usr/local/bin/dockerize

RUN echo "daemon off;" >> /etc/nginx/nginx.conf

ADD default.tmpl /etc/nginx/sites-available/default.tmpl

EXPOSE 80

CMD ["dockerize", "-template", "/etc/nginx/sites-available/default.tmpl:/etc/nginx/sites-available/default", "-stdout", "/var/log/nginx/access.log", "-stderr", "/var/log/nginx/error.log", "nginx"]
