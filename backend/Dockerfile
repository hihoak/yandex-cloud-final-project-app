######## Start a new stage from scratch #######
FROM ubuntu:jammy  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY main .

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./main"] 