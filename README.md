### Build
 docker build -t gigiozzz/producer-http:0.0.1 .

 docker push gigiozzz/producer-http:0.0.1

### Usage on kube
 kubectl -n kafka run kafka-producer -ti --image=gigiozzz/producer-http:0.0.1 --rm=true  --env=KAFKA=my-cluster-kafka-bootstrap:9092