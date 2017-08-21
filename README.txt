- scrapeIDs.txt - contains all of the URLs to scrape from.
- scraper.py - goes through all of the URLs in scrapeIDs.txt and scrapes the text from each and writes them all into separate files.
- svd.go - Uses the text files scraped in the '/extraction' directory to vectorize each document/text file in a term space and writes the vectors to the file 'outputMatrix.txt' by default, or whatever output file you specify by flag.
- cluster.py - Takes 3 parameters when called (threshold value between 0.0 and 2.0, dimensions of each vector in the input file & the name of the input file). Clusters the vectors of the documents in the input file based on cosine similarity between the vectors and the threshold value inputed and marshals cluster results to file 'clusterData.json' with the filenames and relevant keywords (stemmed) in each cluster.
- CalcEnergy.py - Intermediate script I wrote to figure out how many dimensions are necessary to capture 90% of the total energy in the matrix of eigenvalues after SVD.
- ./extraction - Contains text files of the scraped text of the original links from scrapeIDs.txt
- ./deepExtraction - Contains text files of the text scraped from the links on the pages of the links from scrapeIDs.txt. Stopped collecting after about 600 links because it was taking a long time.


- clusterData.json - Contains the results of the clustering of the text of the original links from scrapeIDs.txt based only on TEXT scraped from the links. Contains a json struct with dictionaries corresponding to the IDs of each document in each cluster label, the files in each cluster label and the most important terms in each cluster label based on the highest tf-idf scores in the clusters.


How to Run:
- In order to run the scraper: 
	python scraper.py
- In order to run the Latent Semantic Analysis on the scraped texts: 
	go build
	./CiscoSecurity_Challenge "output file"
- In order to run clustering upon the results from LSA:
	python cluster.py 'threshold value' 'dimensions in each vector' 'input file'

For example I ran the following in the following order and the results are in the respective files:

	python scraper.py

	go build
	./CiscoSecurity_Challenge outputMatrix.txt

	python cluster.py 0.95 34 outputMatrix.txt

Warning: If you do want run the LSA/SVD program on a directory of you choosing, I have done some manual reductions to the matrix in svd.go based on the scraped files that will probably have to be tweaked for a different data set.
