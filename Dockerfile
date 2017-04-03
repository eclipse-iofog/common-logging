FROM jeanblanchard/java:jdk-8u77

RUN mkdir /jar-file
COPY . /jar-file
RUN cd /jar-file
ENV SELFNAME "TEST_LOG_CONTAINER"

CMD ["java", "-jar", "/jar-file/log_container.jar"]