FROM ubuntu:18.04 as runnerContext
RUN echo "test" > test.txt
RUN cat test.txt
COPY dockerBuildContext/entry.bash .
ENTRYPOINT [ "bash", "entry.bash" ]
