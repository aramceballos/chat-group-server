FROM postgres:16-alpine

# Copy the database schema to the container
COPY up.sql /docker-entrypoint-initdb.d/1.sql

CMD ["postgres"]