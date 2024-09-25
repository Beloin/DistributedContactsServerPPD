import socket
import struct
import os
import time

servers = ["contact-server-1:9000", "contact-server-2:9001", "contact-server-3:9002"]

clear = lambda: os.system('clear')

CLIENT_IDENTIFIER = 2  # 1 byte for client identifier
MAX_NAME_LENGTH = 256  # Max name length including null terminator

last_server = None
client_name = "MyName"
def connect_to_server(server):
    host, port = server.split(":")
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.connect((host, int(port)))
        print(f"Connected to {server}")
        global last_server
        last_server = host
        send_initial_conn(s)
        return s
    except socket.error as e:
        print(f"Failed to connect to {server}: {e}")
        s.close()
        return None

def send_initial_conn(sock: socket.socket):
    print(f"Connected to server at {last_server}")
    # Send client identifier
    sock.sendall(bytes([2]))

    # Send a message to the server
    message = create_str(client_name)
    sock.sendall(message)


def safe_send(sock, bytess):
    try:
        sock.sendall(bytess)
    except socket.error as e:
        print(f"Send failed: {e}")
        return False
    return True


def safe_recv(sock, buffer_size=1024):
    try:
        data = sock.recv(buffer_size)
        if not data:
            print("Connection closed by the server")
            return None
        return data
    except socket.error as e:
        print(f"Receive failed: {e}")
        return None


def convert_4bytes_to_uint32(buffer: bytes) -> int:
    # Ensure the buffer is 4 bytes long
    assert len(buffer) == 4, "Buffer must be exactly 4 bytes"

    # Unpack as a 4-byte unsigned integer ('>I' for big-endian, '<I' for little-endian)
    return struct.unpack(">I", buffer)[0]


def create_str(msg: str, lenj=MAX_NAME_LENGTH) -> bytes:
    # Create a message with a 1-byte identifier and 256-byte null-terminated string for the client name
    encoded_name = msg.encode("utf-8")[: lenj - 1] + b"\x00"
    encoded_name = encoded_name.ljust(lenj, b"\x00")  # Pad to make it exactly 256 bytes
    return encoded_name

def parse_null_terminated_string(buffer):
    # Decode the byte buffer to a string and strip the null character '\0' at the end
    string = buffer.decode('utf-8')  # Assuming the buffer is UTF-8 encoded
    return string.split('\0', 1)[0]


def create_contact(sock):
    name = input("Enter contact name: ")
    number = input("Enter contact number: ")

    ok = safe_send(sock, bytes([1]))
    if not ok:
        print("Connection with server ended")
        return False

    name = create_str(name)
    number = create_str(number, 20)

    ok = safe_send(sock, name)
    if not ok:
        print("Connection with server ended")
        return False

    ok = safe_send(sock, number)
    if not ok:
        print("Connection with server ended")
        return False
    
    input("Created! Press <enter> to continue.")

    return True

def delete_contact(sock):
    name = input("Enter contact name to delete: ")
    ok = safe_send(sock, bytes([2]))
    if not ok:
        print("Connection with server ended")
        return False

    name = create_str(name)
    ok = safe_send(sock, name)
    if not ok:
        print("Connection with server ended")
        return False
    
    input("Deleted! Press <enter> to continue.")

    return True


def list_all_contacts(sock):
    ok = safe_send(sock, bytes([3]))
    if not ok:
        return False
    buffer = safe_recv(sock, 4)
    if not buffer:
        return False

    size = convert_4bytes_to_uint32(buffer)
    print("Amount:", size)
    for _ in range(size):
        buffer = safe_recv(sock, 256)
        if not buffer:
            return False
        name = parse_null_terminated_string(buffer)

        buffer = safe_recv(sock, 20)
        if not buffer:
            return False
        number = parse_null_terminated_string(buffer)

        print("\tName: ", name)
        print("\tNumber: ", number)

    input("Press <enter> to continue.")

    return True


def show_menu():
    print("\nMenu:")
    print("1. Create contact")
    print("2. Delete a contact")
    print("3. List all contacts")
    print("4. Exit")


def main():
    global client_name
    client_name = input("Your Name: ")

    # Try to connect to the first available server
    sock = None
    for server in servers:
        sock = connect_to_server(server)
        if sock:
            break

    if not sock:
        print("Could not connect to any server")
        return


    while True:
        clear()
        show_menu()
        choice = input("Choose an option: ")
        
        if choice == '1':
            ok = create_contact(sock)
            if not ok:
                # Retry connection
                for server in servers:
                    sock = connect_to_server(server)
                    if sock:
                        break

                if not sock:
                    print("Could not connect to any server")
                    return

        elif choice == '2':
            ok = delete_contact(sock)
            if not ok:
                # Retry connection
                for server in servers:
                    sock = connect_to_server(server)
                    if sock:
                        break

                if not sock:
                    print("Could not connect to any server")
                    return
        elif choice == '3':
            ok = list_all_contacts(sock)
            if not ok:
                # Retry connection
                for server in servers:
                    sock = connect_to_server(server)
                    if sock:
                        break

                if not sock:
                    print("Could not connect to any server")
                    return
        elif choice == '4':
            print("Exiting program.")
            break
        else:
            print("Invalid option. Please try again.")


if __name__ == "__main__":
    main()
