import socket
import struct

# Constants
SERVER_IP = "127.0.0.1"  # Replace with your server's IP address
SERVER_PORT = 9001  # Replace with your server's port
CLIENT_IDENTIFIER = 2  # 1 byte for client identifier
MAX_NAME_LENGTH = 256  # Max name length including null terminator

servers = [ "contact-server-1:9000", "contact-server-2:9001", "contact-server-3:9002"]


def convert_4bytes_to_uint32(buffer: bytes) -> int:
    # Ensure the buffer is 4 bytes long
    assert len(buffer) == 4, "Buffer must be exactly 4 bytes"

    # Unpack as a 4-byte unsigned integer ('>I' for big-endian, '<I' for little-endian)
    return struct.unpack(">I", buffer)[0]


def create_str(msg: str, lenj=MAX_NAME_LENGTH) -> bytes:
    # Create a message with a 1-byte identifier and 256-byte null-terminated string for the client name
    encoded_name = msg.encode("utf-8")[: lenj - 1] + b"\x00"
    encoded_name = encoded_name.ljust(
        lenj, b"\x00"
    )  # Pad to make it exactly 256 bytes
    return encoded_name


def connect_to_server():
    # Create a socket object
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # Connect to the server
    client_socket.connect((SERVER_IP, SERVER_PORT))
    print(f"Connected to server at {SERVER_IP}:{SERVER_PORT}")

    # Send client identifier
    client_socket.sendall(bytes([2]))

    # Send a message to the server
    client_name = "ClientName"
    message = create_str(client_name)
    client_socket.sendall(message)

    input("paused: new contact")
    client_socket.sendall(bytes([1]))
    name = "Beloin Sena 2.0"
    number = "666-777-999"
    name = create_str(name)
    number = create_str(number, 20)
    print("Len Name", len(name))
    print("Len Number", len(number))
    client_socket.sendall(name)
    client_socket.sendall(number)

    input("paused: List all")
    client_socket.sendall(bytes([3]))

    buffer = client_socket.recv(4)
    size = convert_4bytes_to_uint32(buffer)
    print("Amount:", size)
    for _ in range(size):
        buffer = client_socket.recv(256)
        name = str(buffer)
        buffer = client_socket.recv(20)
        number = str(buffer)

        print("Name: ", name)
        print("Number: ", number)

    input("paused: Delete")
    client_socket.sendall(bytes([2]))

    name = "Beloin Sena 2.0"
    name = create_str(name)
    print("Len Name", len(name))
    client_socket.sendall(name)

    input("paused: List all")
    client_socket.sendall(bytes([3]))

    buffer = client_socket.recv(4)
    size = convert_4bytes_to_uint32(buffer)
    print("Amount:", size)
    for _ in range(size):
        buffer = client_socket.recv(256)
        name = str(buffer)
        buffer = client_socket.recv(20)
        number = str(buffer)

        print("Name: ", name)
        print("Number: ", number)


if __name__ == "__main__":
    connect_to_server()
