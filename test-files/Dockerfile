FROM phusion/baseimage

RUN apt-get update

RUN apt-get install -y golang

RUN sed -i 's/#force_color_prompt=yes/force_color_prompt=yes/g' /root/.bashrc
