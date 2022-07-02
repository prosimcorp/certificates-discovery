########################################################################################################################
# BASE
########################################################################################################################
FROM debian:buster-slim as base

# Update and install base packages if it's necesary
RUN apt-get update --yes && \
    apt-get install --yes \
      ca-certificates

# Prepare app directory
RUN mkdir -p /usr/app/certificates-discovery/
WORKDIR /usr/app/certificates-discovery/

# Configure entrypoint
COPY ./docker-entrypoint.sh /usr/local/bin/
RUN chmod 0775 /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]
# This step will be replaced by the entrypoint plus the args field defined in the kubernetes eployment manifest
CMD ["sh"]

########################################################################################################################
# BUILD
########################################################################################################################
FROM golang:1.17-buster as build
# Update and install base packages if it's necesary
RUN apt-get update --yes

# Copy the application files
COPY . /usr/app/certificates-discovery/

# Build the application
RUN cd /usr/app/certificates-discovery/ \
    && make build

########################################################################################################################
# APPLICATION
# FOR PROD BUILD ADD TARGET FLAG: docker build . --tag 'certificates-discovery:buster-slim' --target application
########################################################################################################################
FROM base as application

# Copy the build application to the working directory
COPY --from=build /usr/app/certificates-discovery/bin/* /usr/app/certificates-discovery/

# Prepare executable permissions
RUN chmod -R 0775 /usr/app/certificates-discovery/certificates-discovery

# Link application certificates-discovery
RUN ln -s /usr/app/certificates-discovery/certificates-discovery /usr/local/bin/certificates-discovery && \
    chmod +x /usr/local/bin/certificates-discovery

########################################################################################################################
# DEBUG
# FOR A DEBUG BUILD: docker build . --tag 'certificates-discovery:buster-slim'
########################################################################################################################
FROM application as debug
# Install debug packages
RUN apt-get update --yes && \
    apt-get install --yes --no-install-recommends \
        bash \
        procps
