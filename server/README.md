# Server
The server of the ramp-up Go challenge servers a key-value store backed database over http.

## Criteria
* use net/http package 
* use map[string]string as “database” 
* serve "database" as a key-value store
* consistent method calls (i.e. use path, query or body params consistently)
* Request content should be verified to match expectations and an appropriate status code should be returned on violations
* Implement the following calls:
  * Set a (1) key to a (1) value.
  * Use the PUT method and read data from the request body, query or path. Return a “success” status code if the operation is successful.
  * Get a key’s value. Use the GET method and write data to the response body. Return the appropriate HTTP status code when the key is not found.
  * Delete a key and its value. Use the DELETE method and return the appropriate HTTP status code when the key is not found.
  * Use the HTTP status code to differentiate between setting (PUT) a new key and updating an existing key.
