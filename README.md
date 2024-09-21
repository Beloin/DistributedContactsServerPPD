# Golang server

# Connection process

## On connect:

1° byte:
  - 1 == "Server"
  - 2 == "Client"  

2-257° byte (n = 256):
  - Server or Client name

## Loop

### Server:

- 1° byte:
  - 1 == "Update"
  - 2 == "HeartBeat"
  - 3 == "UpdateClock" (OPTIONAL)
  - 4 == "AskForUpdate"

1. Update a contact:

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| Clock | 4 | 11 |
| UserName | 256 | "Beloin" |
| ContactName | 256 | "Juan" |
| PhoneNumber | 20 | "85999999999" |

2. Hearbeat

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| HealthStatus | 1 | 1,2,3 |
| Clock | 4 | 11 |

If __My Server's__ Clock >= __Other Server's__ Clock
  - Send "UpdateClock" command

If __My Server's__ Clock < __Other Server's__ Clock
  - Update my clock and keep listening.  

3. UpdateClock (OPTIONAL)

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| Clock | 4 | 11 |

4. AskForUpdate

In this scenario, __My Server__ will send all contacts with the following structure:

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| ContactLen | 4 | 11 |

And then send for each contat a "Update" request:

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| Clock | 4 | 11 |
| UserName | 256 | "Beloin" |
| ContactName | 256 | "Juan" |
| PhoneNumber | 20 | "85999999999" |


### Client

- 1° byte:
  - 1 == "Update"
  - 2 == "ListAll"

1. Update

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| UserName | 256 | "Beloin" |
| ContactName | 256 | "Juan" |
| PhoneNumber | 20 | "85999999999" |

2. ListAll

Server will send:

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| ContactLen | 4 | 11 |

And then send for each contact an "Update" request:

| Name | Bytes | Example |
| --------------- | --------------- | --------------- |
| UserName | 256 | "Beloin" |
| ContactName | 256 | "Juan" |
| PhoneNumber | 20 | "85999999999" |


# Resources:

1. [Limit](https://mostafa.dev/why-do-tcp-connections-in-go-get-stuck-reading-large-amounts-of-data-f490a26a605e) TCP reading.
