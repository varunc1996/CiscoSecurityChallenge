- scrapeIDs.txt - contains all of the URLs to scrape from.
- scraper.py - goes through all of the URLs in scrapeIDs.txt and scrapes the text from each and writes them all into separate files.
- svd.go - Uses the text files scraped in the '/extraction' directory to vectorize each document/text file in a term space and writes the vectors to the file 'outputMatrix.txt'
- cluster.py - Takes 1 parameter when called (threshold value between 0.0 and 2.0). Clusters the vectors of the documents in outputMatrix.txt based on cosine similarity between the vectors and the threshold value inputed and marshals cluster results to file 'clusterData.json' with the filenames and relevant keywords (stemmed) in each cluster.
- CalcEnergy.py - Intermediate script I wrote to figure out how many dimensions are necessary to capture 90% of the total energy in the matrix of eigenvalues after SVD.
- ./extraction - Contains text files of the scraped text of the original links from scrapeIDs.txt
- ./deepExtraction - Contains text files of the text scraped from the links on the pages of the links from scrapeIDs.txt. Stopped collecting after about 600 links because it was taking a long time.


- clusterData.json - Contains the results of the clustering of the text of the original links from scrapeIDs.txt based only on TEXT scraped from the links. Contains a json struct with dictionaries corresponding to the IDs of each document in each cluster label, the files in each cluster label and the most important terms in each cluster label based on the highest tf-idf scores in the clusters.