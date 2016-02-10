CC = gcc
CFLAGS = -c -Wall -Wextra -std=c99 -I.
LDFLAGS = -static

CFLAGS += $(shell pkg-config --cflags jansson)
LDFLAGS += $(shell pkg-config --libs jansson)

.PHONY: clean

print-metadata: main.o pwd_ng.o
	$(CC) -o print-metadata pwd_ng.o main.o $(LDFLAGS)

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

clean:
	rm print-metadata main.o
