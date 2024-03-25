FROM mcr.microsoft.com/devcontainers/base:bullseye

# Install MySql for Go backend unit tests. The tests expect MySql to be
# installed directly in the development environment rather than in another
# docker continer such as the dev database used for manual testing.
COPY install-mysql.sh .
RUN chmod +x ./install-mysql.sh && ./install-mysql.sh
