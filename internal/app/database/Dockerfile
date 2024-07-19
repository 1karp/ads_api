# Use the official PostgreSQL 13 image as the base
FROM postgres:13-alpine

# Set environment variables
ENV POSTGRES_USER=os.Getenv("POSTGRES_USER")
ENV POSTGRES_PASSWORD=os.Getenv("POSTGRES_PASSWORD")
ENV POSTGRES_DB=os.Getenv("POSTGRES_DB")

# Set up the volume for persistent data
VOLUME /var/lib/postgresql/data

# Expose the default PostgreSQL port
EXPOSE 5432

# The CMD instruction is not needed as it's inherited from the base image