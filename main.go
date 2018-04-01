package main

import (
	"strings"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"regexp"
	"github.com/mikemintang/go-curl"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/go-gomail/gomail"
	"math"
)


//here is config
var the_url_indeed = "https://de.indeed.com/jobs?"

//hier schreiben Ihre traurige Beruf
var the_job_you_search_for = ""

//hier schreiben Ihre Stadt
var the_city = ""

//email_setting
//Ihre Email Addresse
var my_email = ""

//Ihre Email Smtp Server z.B: smtp.gmail.com
var smtp_host = ""

//Ihre Email Smtp Port z.B: 587  <- Gmail smtp
var smtp_port = 25 //normalerweise ist 25

//Ihr email username für loggen
var email_username = ""

//Ihre email pasword für loggen
var email_password = ""

//Email Title z.B: Fachinformatiker Bewerbung
var email_subject = ""

//Email Body,  \n bedeutet nächste linie
var email_body =
	"Sehr geehrte Damen und Herren,\n" +
	"........." +
	"Mit freundlichen Grüßen."

func main() {

	os.Create("emails.txt")

	count := get_count(the_url_indeed, the_job_you_search_for, the_city)

	var total_pages = int(math.Ceil(float64(count) / float64(14)))

	//var link string

	for i := 1; i <= total_pages; i++{
		get_link(the_url_indeed, the_job_you_search_for, the_city, i)
	}
	fmt.Println("\nGet email address completed\n")

	check_email()

	fmt.Println("\nReset emails.txt completed\n")

	email_send()

}


func get_count(url string,job string,ort string)(page int){
	job = strings.Replace(job," ","+", -1)
	ort = strings.Replace(ort," ","+", -1)
	ganz_url := url + "q=" + job + "&l=" + ort

	req, err:= goquery.NewDocument(ganz_url)

	if(err != nil){
		log.Fatal(err)
	}

	page_count_in_string := req.Find("#searchCount").Text()

	i := 1
	var last_word string
	//var total_string string

	for{
		last_word = page_count_in_string[len(page_count_in_string) - i:]

		//fix the string bug .
		last_word = strings.Replace(last_word,".","", -1)

		if(string([]rune(last_word)[0]) == " "){
			break
		}
		i = i + 1
	}

	last_word = strings.Replace(last_word," ","",-1)

	the_page, the_err := strconv.Atoi(last_word)

	if(the_err != nil){
		log.Fatal(the_err)
	}

	return the_page
}


func get_link(url string,job string,ort string,page int){
	job = strings.Replace(job," ","+", -1)
	ort = strings.Replace(ort," ","+", -1)
	the_page := strconv.Itoa(page)

	ganz_url := url + "q=" + job + "&l=" + ort + "&start=" + the_page

	req, err := goquery.NewDocument(ganz_url)

	if(err != nil){
		log.Fatal(err)
	}


	req.Find("a[rel]").Each(func(i int, selection *goquery.Selection) {
		the_title, _ := selection.Attr("rel")
		if(the_title == "noopener nofollow"){
			href , _:= selection.Attr("href")
			get_check := get_email("https://de.indeed.com" + href)

			if(get_check == "error"){

			}else{
				fd,_:=os.OpenFile("emails.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
				buf:=[]byte(get_check + "\n")
				fd.Write(buf)
				fd.Close()
				fmt.Println(get_check)
			}
		}
	})
}

func get_email(link string)(email string){
	//This function will try to find email address from page and return the email address
	the_email_should := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(the_email_should)

	req := curl.NewRequest()
	resp, err := req.
		SetUrl(link).
		Get()

	var the_return_email string

	if err != nil {
		fmt.Println(err)
	} else {
		the_email := reg.FindString(resp.Body)
		if(the_email != ""){
			the_return_email = the_email
		}else{
			the_return_email = "error"
		}
	}
	return the_return_email
}

func check_email(){
	//This function will check all email address and remove repeated one.
	buf, _ := ioutil.ReadFile("emails.txt")

	emails_split := strings.Split(string(buf),"\n")

	var new_email_list []string

	//create new emails.txt
	os.Create("emails.txt")

	//write new data in emails.txt
	fd,_:=os.OpenFile("emails.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)


	for i:= range emails_split{
		if(in_array(emails_split[i],new_email_list) == false){
			new_email_list = append(new_email_list ,emails_split[i])
		}
	}

	for i := range new_email_list{
		buf:=[]byte(new_email_list[i] + "\n")
		fd.Write(buf)
	}

	fd.Close()
	fmt.Println("Ok")
}

func email_send(){
	buf, _ := ioutil.ReadFile("emails.txt")

	target_emails := strings.Split(string(buf),"\n")

	m := gomail.NewMessage()

	m.SetAddressHeader("From",my_email,"")

	m.SetHeader("Subject", email_subject)

	m.SetBody("text",email_body)

	skillfolder := "send_together/"

	files, _ := ioutil.ReadDir(skillfolder)

	for _,file := range files {
		if file.IsDir() {
			continue
		} else {
			m.Attach(skillfolder + file.Name())
		}
	}


	for i:= range target_emails{
		if(target_emails[i] == ""){
			continue
		}
		m.SetHeader("To", m.FormatAddress(target_emails[i], ""))

		d := gomail.NewPlainDialer(smtp_host, smtp_port, email_username, email_password)

		if err := d.DialAndSend(m); err != nil {
			fmt.Println("\n Send to:" + target_emails[i] + " failed")
			continue
		}
		log.Println("\n Send to: " + target_emails[i] + " succeed \n")
	}
}


func in_array(str string, the_array []string)(bool){
	for i:= range the_array{
		if(str == the_array[i]){
			return true
		}
	}
	return false
}
