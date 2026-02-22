curl -sX POST http://localhost/apply \
  -H "Authorization: Bearer BINGO" \
  -H "Content-Type: application/yaml" \
  --data-binary @job.yml
