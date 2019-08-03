node {
    git url: 'https://github.com/usrpro/file-handling-api.git'
    sh 'printenv | more'
    docker.image('postgres:latest').withRun('-e "POSTGRES_DB=s3db_01"') { c ->
        docker.image('postgres:latest').inside("--link ${c.id}:db") {
            /*  */
            echo 'Postgresql checkpoint passed.'
        }
        docker.image('golang:latest').inside("--link ${c.id}:db") {
                    
                    sh 'go version'
                    sh 'go get github.com/anthonynsimon/bild/imgio'
                    sh 'go get github.com/anthonynsimon/bild/transform' 
                    sh 'go get github.com/fulldump/goconfig' 
                    sh 'go get github.com/jackc/pgx' 
                    sh 'go get github.com/minio/minio-go' 
                    sh 'go get gopkg.in/inconshreveable/log15.v2'
            withEnv(['S3_HOST=play.minio.io:9000',
            'S3_KEY=Q3AM3UQ867SPQQA43P2F',
            'S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG',
            'S3_BUCKET=magick-crop',
            'DATABASE_HOST=db']) {
                sh 'go build'
                sh 'go test'
                }
        }
    }
}      