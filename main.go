package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type records struct {
	hasRecord bool;
	record string;
}

func main() {

	scanner := bufio.NewScanner(os.Stdin);

	fmt.Printf("domain, hasMX, hasSPF, sprRecord, hasDMARC, dmarcRecord\n");

	fmt.Println("Please enter domain or email: \n");

	for scanner.Scan() {
	
		if scanner.Text() == "" {
			log.Fatal("Please submit domain");
		}
		checkDomain(scanner.Text());
	}

	err := scanner.Err();

	if err != nil {
		log.Fatal("Error: could not read from input: %v\n", err);
	}
}

func lookupMxRecord(mxRecordChannel chan<- records, domain string) {

	fmt.Println("Looking up MX record...");

	mxRecords, err := net.LookupMX(domain);

	if err != nil {
		log.Printf("Error: %v\n", err);
	}

	if len(mxRecords) > 0 {
		var record records;
		record.hasRecord = true;
		mxRecordChannel <- record;
	}

	defer close(mxRecordChannel);
}

func lookupSpfRecord(spfRecordChannel chan<- records, domain string) {

	fmt.Println("Looking up SPF record...");

	txtRecords, err := net.LookupTXT(domain);

	if err != nil {
		log.Printf("Error: %v\n", err);
	}

	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=spf1") {
			var spfRecord records;
			spfRecord.hasRecord = true;
			spfRecord.record = record;
			spfRecordChannel <- spfRecord;
		}
	}

	defer close(spfRecordChannel);
}

func lookupDmarcRecord(dmarcRecordChannel chan<- records, domain string) {
 
	fmt.Println("Looking up DMARC record...");

	dmarcRecords, err := net.LookupTXT("_dmarc." + domain);

	if err != nil {
		log.Printf("Error: %v\n", err);
	}

	for _, record := range dmarcRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			var dmarcRecord records;
			dmarcRecord.hasRecord = true;
			dmarcRecord.record = record
			dmarcRecordChannel <- dmarcRecord;
		}
	}

	defer close(dmarcRecordChannel);
}

func checkDomain(domain string) {

	mxRecordChannel := make(chan records, 1);
	spfRecordChannel := make(chan records, 1);
	dmarcRecordChannel := make(chan records, 1);

	go lookupMxRecord(mxRecordChannel, domain);

	go lookupSpfRecord(spfRecordChannel, domain);

	go lookupDmarcRecord(dmarcRecordChannel, domain);

	mxRecord := <- mxRecordChannel;
	spfRecord := <- spfRecordChannel;
	dmarcRecord := <- dmarcRecordChannel;
	
	fmt.Printf("%v, %v, %v, %v, %v, %v", domain, mxRecord.hasRecord, spfRecord.hasRecord, spfRecord.record, dmarcRecord.hasRecord, dmarcRecord.record);

	fmt.Println("\nDomain lookup successful.\n");

}