from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import datetime

# Define a custom HTTP request handler
class RequestHandler(BaseHTTPRequestHandler):
    # Set the necessary headers for the response
    def _set_headers(self):
        self.send_header('Access-Control-Allow-Origin', 'http://localhost:8080') # * for all, http://localhost:8080 for golang server
        # Allow the specified methods
        self.send_header('Access-Control-Allow-Methods', 'POST')
        # Allow the specified headers
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')

    # Handle OPTIONS requests
    def do_OPTIONS(self):
        # Send a 200 OK response
        self.send_response(200)
        # Set the headers
        self._set_headers()
        # End the headers
        self.end_headers()

    # Handle POST requests
    def do_POST(self):
        # If the path is /save_chat
        if self.path == "/save_chat":
            # Get the length of the content
            content_length = int(self.headers['Content-Length'])
            # Read the content
            post_data = self.rfile.read(content_length)
            # Parse the content as JSON
            chat_data = json.loads(post_data)
            
            # Get the current time and format it
            now = datetime.datetime.now()
            filename = now.strftime("websocket_chat_history_%Y_%m_%d_%H_%M_%S.txt")
            
            # Open the file and write the chat data to it
            with open(filename, "w", encoding="utf-8") as f:
                for message in chat_data:
                    line = f"{message['username']}: {message['content']}\n"
                    f.write(line)
            
            # Send a 200 OK response
            self.send_response(200)
            # Set the headers
            self._set_headers()
            # End the headers
            self.end_headers()
            # Write the response body
            self.wfile.write(b"Chat data saved successfully!")
        else:
            # If the path is not /save_chat, send a 404 Not Found response
            self.send_response(404)
            self._set_headers()
            self.end_headers()

# Function to start the server
def run(server_class=HTTPServer, handler_class=RequestHandler, port=8000):
    server_address = ('', port)
    http = server_class(server_address, handler_class)
    print(f'Starting http server on port {port}')
    http.serve_forever()

# If this script is run directly, start the server
if __name__ == "__main__":
    run()