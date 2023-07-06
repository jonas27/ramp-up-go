# Ramp-up challenge
This is a simple http server and client for a ramp up challenge.
For information on the challenge see [here.](https://docs.google.com/document/d/1BtVU34iuoQEs9B9N6QOl20nF1WM_e_OOyWrm1eelf-s/edit#heading=h.rxmn8ufj7ae2)

## Server
The server serves a key-value store backed database over http.
See criteria below.

### Criteria
* use net/http package 
* use map[string]string as “database” 
* serve "database" as a key-value store
* consistent method calls (i.e. use path, query or body params consistently)
* Request content should be verified to match expectations and an appropriate status code should be returned on violations
* Implement the following calls:
  * (✓) Set a (1) key to a (1) value.
  * (✓) Use the PUT method and read data from the request body, query or path. Return a “success” status code if the operation is successful.
  * (✓) Get a key’s value. Use the GET method and write data to the response body. Return the appropriate HTTP status code when the key is not found.
  * (✓) Delete a key and its value. Use the DELETE method and return the appropriate HTTP status code when the key is not found.
  * (✓) Use the HTTP status code to differentiate between setting (PUT) a new key and updating an existing key.
  * (✓) Bonus: POST

## Client
The client interacts with the server through http requests.
See criteria below.

### Criteria
* client can set, get and delete keys and values by communicating with the server
* Make use of the flag package (why flags? cmd line tools conda/kong)
* print error messages to stderr
Example usage: myclient -m=put --key=foo --value=bar
  * (✓) Print an existing key to stdout (GET)
  * (✓) Tell users when a key was created or updated (PUT)
  * (✓) Tell users when a key was successfully deleted (DELETE)



