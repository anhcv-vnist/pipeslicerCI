curl -X POST "http://127.0.0.1:3000/pipelines/build" \
     -H "Content-Type: multipart/form-data" \
     -F "url=git@github.com:vanhcao3/pipeslicerCI.git" \
     -F "branch=main" \
     -F "file=@pipeslicer-ci.yaml"
