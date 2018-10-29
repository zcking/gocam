# GoCam
GoCam is a free, open-source security camera software solution. The intent of
GoCam is to provide a more accessible, resilient solution to video
surveillance.

I run GoCam on a Raspberry Pi 3 Model B with an attached NO-IR Camera module. For a
full shopping list of parts, see below (prices subject to change):
* [Raspbery Pi 3 Model B](https://www.amazon.com/gp/product/B01CD5VC92/ref=oh_aui_detailpage_o00_s00?ie=UTF8&psc=1) ($37.00)
* [NO-IR Camera Module](https://www.amazon.com/gp/product/B01ER2SMHY/ref=oh_aui_detailpage_o01_s00?ie=UTF8&psc=1) ($25.00)
* [Micro SD Card (32 GB)](https://www.amazon.com/Sandisk-Ultra-Micro-UHS-I-Adapter/dp/B073JWXGNT/ref=pd_cp_147_1?_encoding=UTF8&pd_rd_i=B073JWXGNT&pd_rd_r=d4f96f46-c38a-11e8-bcaa-d5dbd61d2792&pd_rd_w=7Lria&pd_rd_wg=uusOo&pf_rd_i=desktop-dp-sims&pf_rd_m=ATVPDKIKX0DER&pf_rd_p=3f5155f5-5438-4fc3-ab83-bf013cfc8883&pf_rd_r=JJW0VA0VNHPM7FT1EFE8&pf_rd_s=desktop-dp-sims&pf_rd_t=40701&psc=1&refRID=JJW0VA0VNHPM7FT1EFE8) ($9.00)
* [Micro USB 5V 2.5A Power Supply](https://www.amazon.com/gp/slredirect/picassoRedirect.html/ref=pa_sp_atf_aps_sr_pg1_2?ie=UTF8&adId=A01621102WG9ZU4WFVYNE&url=https%3A%2F%2Fwww.amazon.com%2FEnokay-Supply-Raspberry-Charger-Adapter%2Fdp%2FB01MZX466R%2Fref%3Dsr_1_2_sspa%3Fie%3DUTF8%26qid%3D1538186236%26sr%3D8-2-spons%26keywords%3Draspberry%2Bpi%2B3%2Bofficial%2Bpower%2Bsupply%26psc%3D1&qualifier=1538186236&id=8507541985793408&widgetName=sp_atf) ($8.00)
  
 
Total Cost: $79.00 (excluding taxes)

**Note:** I chose to use a No-IR camera simply because they are cheap and operate in both day and night
(although daytime colors tend to be a bit washed out), and they do not require any additional
filters or special lenses.

---

## Running GoCam
Once you have a binary of GoCam to run, you 
can start it like any other program, 
by simply executing `./gocam`.  

Note: GoCam expects a YAML configuration file 
at `./config/default.yaml` by default. You may 
refer to the configuration file in this repository 
for an example.

---

## Building
Currently, this project is an infant and I haven't set up any Docker images, Vagrant files, or anything of the sort.
However, if you have golang installed, compiling is very simple: `go build`.

If you would like to build the application for Raspbian (OS), please note you must already be on an ARM
architecture to compile it, since this application requires the CGO bindings due to the GoCV dependency.
For now, building for raspberry pi is quite tedious as it requires building Go 1.11 and OpenCV from source:  

1. Install legacy Golang (required to compile Golang 1.11)
```bash
apt update -y
apt-get install golang
```

2. Build Golang 1.11 from source
```bash
git clone https://go.googlesource.com/go
cd go
git checkout go1.11
cd src
./all.bash
cd ../..
mv go /usr/local/go
```
3. Add Go 1.11 to system `PATH`
```bash
echo -e export PATH="\$PATH:/usr/local/go/bin" >> /etc/profile
export PATH="$PATH:/usr/local/go/bin"
```
4. Set your `GOPATH` system environment variable
```bash
echo -e export GOPATH=$HOME >> ~/.profile
export GOPATH=$HOME
```
5. Build OpenCV from source
```bash
go get -u -d gocv.io/x/gocv
cd $GOPATH/src/gocv.io/x/gocv
make install
```
6. Finally...build GoCam!
```bash
go install github.com/zcking/gocam
```

---

## Remote Access: Networking Setup
To allow you to access the web server on the GoCam, 
you'll need to modify your firewall. You can do 
this with traditional `iptables` like so:  
```bash
# This will allow TCP traffic to port 80 
# on the PI, from anywhere.
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
```

If you don't need SSH access to the GoCam, 
after creating the previous rule, you can 
restrict access to prevent any unauthorized 
access with the following:
```bash
# This will drop all other input traffic
sudo iptables -A INPUT -j DROP
```

To make the above `iptables` changes 
persistent on reboots, you can run the 
following:  
```bash
sudo /sbin/iptables-save
```