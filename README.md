# goApiServer
server written in Go that run small API call


Created a file, shared files
Here's a single-file Go server using only the standard library (needs Go 1.22+ for the GET /path routing syntax).

Run it:

bash
go mod init smallserver   # first time only
go run main.go



Try it:

bash
curl http://localhost:8080/health
curl http://localhost:8080/api/repo
/api/repo calls the GitHub API from the server and returns a trimmed JSON result (name, description, stars, language). It has a 5s timeout, returns a 502 if the upstream fails, and shuts down gracefully on Ctrl+C. The port defaults to 8080 and can be changed with PORT=9000 go run main.go.

If you meant something different, such as calling a specific API of yours, or accepting a parameter like /api/repo?name=owner/repo, tell me which API and I'll adapt it.
