package utils

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//go:embed *.txt
var f embed.FS

// SupportsHugePages
func SupportsHugePages() (bool, error) {
	sc, err := exec.Command("sysctl", "machdep.cpu.extfeatures").Output()
	if err != nil {
		return false, err
	}
	supports := strings.Contains(string(sc), "1GBPAGE")
	return supports, nil
}

// GenerateMACAddress
func GenerateMACAddress() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	// Set the local bit
	buf[0] |= 2
	mac := fmt.Sprintf("56:%02x:%02x:%02x:%02x:%02x", buf[1], buf[2], buf[3], buf[4], buf[5])

	return mac, nil
}

// Retry retries a function
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			fmt.Printf("\r%s", strings.Repeat(".", i))
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

type Protocol int

const (
	Tcp Protocol = iota
	Udp
)

type PortMap struct {
	Host  int
	Guest int
	Proto Protocol
}

// Parses port mapping configurations
func ParsePort(ports string) ([]PortMap, error) {
	var maps []PortMap = nil
	if ports == "" {
		return maps, nil
	}
	mapcount := strings.Count(ports, ",") + 1
	if mapcount > 65535 {
		return nil, errors.New("too many port mappings specified, likely an error. Check config.yaml")
	}
	maps = make([]PortMap, mapcount)
	for i, p := range strings.Split(ports, ",") {
		newmap := PortMap{0, 0, Tcp}
		var herr, gerr error
		if strings.HasSuffix(p, "u") {
			newmap.Proto = Udp
			p = strings.TrimSuffix(p, "u")
		}
		if strings.Contains(p, ":") {
			pair := strings.Split(p, ":")
			if len(pair) != 2 {
				return nil, errors.New("incorrect port mapping pair specified. Check config.yaml")
			}
			newmap.Host, herr = strconv.Atoi(pair[0])
			newmap.Guest, gerr = strconv.Atoi(pair[1])
		} else {
			newmap.Host, herr = strconv.Atoi(p)
			newmap.Guest = newmap.Host
		}
		if herr != nil || gerr != nil {
			return nil, errors.New("error parsing specified ports. Check config.yaml")
		}
		if newmap.Host < 0 || newmap.Host > 65535 || newmap.Guest < 0 || newmap.Guest > 65535 {
			return nil, errors.New("invalid specified ports (must be 0-65535). Check config.yaml")
		}
		maps[i] = newmap
	}
	return maps, nil
}

// Ping checks if connection is reachable
func Ping(ip string, port string) error {
	address, err := net.ResolveTCPAddr("tcp", ip+":"+port)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil
	}

	if conn != nil {
		defer conn.Close()
		return errors.New("port " + port + " already assigned on host")
	}

	return err
}

// StringSliceContains check if string value is in []string
func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Uncompress uncompresses gzip
func Uncompress(source string, destination string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	gzRead, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzRead.Close()

	tarRead := tar.NewReader(gzRead)
	for {
		cur, err := tarRead.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if strings.Contains(string(cur.Name), "..") {
			return fmt.Errorf("archive contains invalid filename: %s", cur.Name)
		}

		os.MkdirAll(destination, 0777)

		switch cur.Typeflag {

		case tar.TypeReg:
			create, err := os.Create(filepath.Join(destination, cur.Name))
			if err != nil {
				return err
			}
			defer create.Close()
			create.ReadFrom(tarRead)
		case tar.TypeLink:
			os.Link(cur.Linkname, cur.Name)
		}
	}
	return nil
}

// Compress creates a tar.gz of a Directory
func Compress(files []string, buf io.Writer) error {
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

// CopyFile copies file from src to dst
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rretrieving image... %3dMB complete", wc.Total/1000000)
}

func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("requested image download is not supported: StatusCode " + strconv.Itoa(resp.StatusCode))
	}

	defer resp.Body.Close()

	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	fmt.Print("\n")
	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}

func GetImageURL(version string) string {
	url := "https://github.com/beringresearch/macpine/releases/download/v.01/" + version
	return url
}

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func GenerateRandomAlias() string {
	var alias string
	var adjectivesString []string
	var nounsString []string

	adjectives, _ := f.ReadFile("adjectives.txt")
	adjectivesString = strings.Split(string(adjectives), "\n")

	nouns, _ := f.ReadFile("nouns.txt")
	nounsString = strings.Split(string(nouns), "\n")

	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(adjectivesString)
	a := adjectivesString[n]

	rand.Seed(time.Now().Unix())
	n = rand.Int() % len(nounsString)
	o := nounsString[n]

	alias = a + "-" + o
	return alias

}

type CmdResult struct {
	Name string
	Err  error
}

func PassphrasePromptForDecryption() (string, error) {
	pass, err := readSecret("enter passphrase:")
	if err != nil {
		return "", fmt.Errorf("could not read passphrase: %v", err)
	}
	p := string(pass)
	return p, nil
}

func PassphrasePromptForEncryption() (string, error) {
	pass, err := readSecret("enter passphrase (leave empty to autogenerate a secure one):")
	if err != nil {
		return "", fmt.Errorf("could not read passphrase: %v", err)
	}
	p := string(pass)
	if p == "" {
		var words []string
		for i := 0; i < 10; i++ {
			rword, err := randomWord()
			if err != nil {
				return "", err
			}
			words = append(words, rword)
		}
		p = strings.Join(words, "-")
		err := printfToTerminal("using autogenerated passphrase %q", p)
		if err != nil {
			return "", fmt.Errorf("could not print passphrase: %v", err)
		}
	} else {
		confirm, err := readSecret("confirm passphrase:")
		if err != nil {
			return "", fmt.Errorf("could not read passphrase: %v", err)
		}
		if string(confirm) != p {
			return "", fmt.Errorf("passphrases didn't match")
		}
	}
	return p, nil
}

func randomWord() (string, error) {
	buf := make([]byte, 2)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	n := binary.BigEndian.Uint16(buf)
	return wordlist[int(n)%2048], nil
}

// wordlist is the BIP39 list of 2048 english words, and it's used to generate
// the suggested passphrases.
var wordlist = strings.Split(`abandon ability able about above absent absorb abstract absurd abuse access accident account accuse achieve acid acoustic acquire across act action actor actress actual adapt add addict address adjust admit adult advance advice aerobic affair afford afraid again age agent agree ahead aim air airport aisle alarm album alcohol alert alien all alley allow almost alone alpha already also alter always amateur amazing among amount amused analyst anchor ancient anger angle angry animal ankle announce annual another answer antenna antique anxiety any apart apology appear apple approve april arch arctic area arena argue arm armed armor army around arrange arrest arrive arrow art artefact artist artwork ask aspect assault asset assist assume asthma athlete atom attack attend attitude attract auction audit august aunt author auto autumn average avocado avoid awake aware away awesome awful awkward axis baby bachelor bacon badge bag balance balcony ball bamboo banana banner bar barely bargain barrel base basic basket battle beach bean beauty because become beef before begin behave behind believe below belt bench benefit best betray better between beyond bicycle bid bike bind biology bird birth bitter black blade blame blanket blast bleak bless blind blood blossom blouse blue blur blush board boat body boil bomb bone bonus book boost border boring borrow boss bottom bounce box boy bracket brain brand brass brave bread breeze brick bridge brief bright bring brisk broccoli broken bronze broom brother brown brush bubble buddy budget buffalo build bulb bulk bullet bundle bunker burden burger burst bus business busy butter buyer buzz cabbage cabin cable cactus cage cake call calm camera camp can canal cancel candy cannon canoe canvas canyon capable capital captain car carbon card cargo carpet carry cart case cash casino castle casual cat catalog catch category cattle caught cause caution cave ceiling celery cement census century cereal certain chair chalk champion change chaos chapter charge chase chat cheap check cheese chef cherry chest chicken chief child chimney choice choose chronic chuckle chunk churn cigar cinnamon circle citizen city civil claim clap clarify claw clay clean clerk clever click client cliff climb clinic clip clock clog close cloth cloud clown club clump cluster clutch coach coast coconut code coffee coil coin collect color column combine come comfort comic common company concert conduct confirm congress connect consider control convince cook cool copper copy coral core corn correct cost cotton couch country couple course cousin cover coyote crack cradle craft cram crane crash crater crawl crazy cream credit creek crew cricket crime crisp critic crop cross crouch crowd crucial cruel cruise crumble crunch crush cry crystal cube culture cup cupboard curious current curtain curve cushion custom cute cycle dad damage damp dance danger daring dash daughter dawn day deal debate debris decade december decide decline decorate decrease deer defense define defy degree delay deliver demand demise denial dentist deny depart depend deposit depth deputy derive describe desert design desk despair destroy detail detect develop device devote diagram dial diamond diary dice diesel diet differ digital dignity dilemma dinner dinosaur direct dirt disagree discover disease dish dismiss disorder display distance divert divide divorce dizzy doctor document dog doll dolphin domain donate donkey donor door dose double dove draft dragon drama drastic draw dream dress drift drill drink drip drive drop drum dry duck dumb dune during dust dutch duty dwarf dynamic eager eagle early earn earth easily east easy echo ecology economy edge edit educate effort egg eight either elbow elder electric elegant element elephant elevator elite else embark embody embrace emerge emotion employ empower empty enable enact end endless endorse enemy energy enforce engage engine enhance enjoy enlist enough enrich enroll ensure enter entire entry envelope episode equal equip era erase erode erosion error erupt escape essay essence estate eternal ethics evidence evil evoke evolve exact example excess exchange excite exclude excuse execute exercise exhaust exhibit exile exist exit exotic expand expect expire explain expose express extend extra eye eyebrow fabric face faculty fade faint faith fall false fame family famous fan fancy fantasy farm fashion fat fatal father fatigue fault favorite feature february federal fee feed feel female fence festival fetch fever few fiber fiction field figure file film filter final find fine finger finish fire firm first fiscal fish fit fitness fix flag flame flash flat flavor flee flight flip float flock floor flower fluid flush fly foam focus fog foil fold follow food foot force forest forget fork fortune forum forward fossil foster found fox fragile frame frequent fresh friend fringe frog front frost frown frozen fruit fuel fun funny furnace fury future gadget gain galaxy gallery game gap garage garbage garden garlic garment gas gasp gate gather gauge gaze general genius genre gentle genuine gesture ghost giant gift giggle ginger giraffe girl give glad glance glare glass glide glimpse globe gloom glory glove glow glue goat goddess gold good goose gorilla gospel gossip govern gown grab grace grain grant grape grass gravity great green grid grief grit grocery group grow grunt guard guess guide guilt guitar gun gym habit hair half hammer hamster hand happy harbor hard harsh harvest hat have hawk hazard head health heart heavy hedgehog height hello helmet help hen hero hidden high hill hint hip hire history hobby hockey hold hole holiday hollow home honey hood hope horn horror horse hospital host hotel hour hover hub huge human humble humor hundred hungry hunt hurdle hurry hurt husband hybrid ice icon idea identify idle ignore ill illegal illness image imitate immense immune impact impose improve impulse inch include income increase index indicate indoor industry infant inflict inform inhale inherit initial inject injury inmate inner innocent input inquiry insane insect inside inspire install intact interest into invest invite involve iron island isolate issue item ivory jacket jaguar jar jazz jealous jeans jelly jewel job join joke journey joy judge juice jump jungle junior junk just kangaroo keen keep ketchup key kick kid kidney kind kingdom kiss kit kitchen kite kitten kiwi knee knife knock know lab label labor ladder lady lake lamp language laptop large later latin laugh laundry lava law lawn lawsuit layer lazy leader leaf learn leave lecture left leg legal legend leisure lemon lend length lens leopard lesson letter level liar liberty library license life lift light like limb limit link lion liquid list little live lizard load loan lobster local lock logic lonely long loop lottery loud lounge love loyal lucky luggage lumber lunar lunch luxury lyrics machine mad magic magnet maid mail main major make mammal man manage mandate mango mansion manual maple marble march margin marine market marriage mask mass master match material math matrix matter maximum maze meadow mean measure meat mechanic medal media melody melt member memory mention menu mercy merge merit merry mesh message metal method middle midnight milk million mimic mind minimum minor minute miracle mirror misery miss mistake mix mixed mixture mobile model modify mom moment monitor monkey monster month moon moral more morning mosquito mother motion motor mountain mouse move movie much muffin mule multiply muscle museum mushroom music must mutual myself mystery myth naive name napkin narrow nasty nation nature near neck need negative neglect neither nephew nerve nest net network neutral never news next nice night noble noise nominee noodle normal north nose notable note nothing notice novel now nuclear number nurse nut oak obey object oblige obscure observe obtain obvious occur ocean october odor off offer office often oil okay old olive olympic omit once one onion online only open opera opinion oppose option orange orbit orchard order ordinary organ orient original orphan ostrich other outdoor outer output outside oval oven over own owner oxygen oyster ozone pact paddle page pair palace palm panda panel panic panther paper parade parent park parrot party pass patch path patient patrol pattern pause pave payment peace peanut pear peasant pelican pen penalty pencil people pepper perfect permit person pet phone photo phrase physical piano picnic picture piece pig pigeon pill pilot pink pioneer pipe pistol pitch pizza place planet plastic plate play please pledge pluck plug plunge poem poet point polar pole police pond pony pool popular portion position possible post potato pottery poverty powder power practice praise predict prefer prepare present pretty prevent price pride primary print priority prison private prize problem process produce profit program project promote proof property prosper protect proud provide public pudding pull pulp pulse pumpkin punch pupil puppy purchase purity purpose purse push put puzzle pyramid quality quantum quarter question quick quit quiz quote rabbit raccoon race rack radar radio rail rain raise rally ramp ranch random range rapid rare rate rather raven raw razor ready real reason rebel rebuild recall receive recipe record recycle reduce reflect reform refuse region regret regular reject relax release relief rely remain remember remind remove render renew rent reopen repair repeat replace report require rescue resemble resist resource response result retire retreat return reunion reveal review reward rhythm rib ribbon rice rich ride ridge rifle right rigid ring riot ripple risk ritual rival river road roast robot robust rocket romance roof rookie room rose rotate rough round route royal rubber rude rug rule run runway rural sad saddle sadness safe sail salad salmon salon salt salute same sample sand satisfy satoshi sauce sausage save say scale scan scare scatter scene scheme school science scissors scorpion scout scrap screen script scrub sea search season seat second secret section security seed seek segment select sell seminar senior sense sentence series service session settle setup seven shadow shaft shallow share shed shell sheriff shield shift shine ship shiver shock shoe shoot shop short shoulder shove shrimp shrug shuffle shy sibling sick side siege sight sign silent silk silly silver similar simple since sing siren sister situate six size skate sketch ski skill skin skirt skull slab slam sleep slender slice slide slight slim slogan slot slow slush small smart smile smoke smooth snack snake snap sniff snow soap soccer social sock soda soft solar soldier solid solution solve someone song soon sorry sort soul sound soup source south space spare spatial spawn speak special speed spell spend sphere spice spider spike spin spirit split spoil sponsor spoon sport spot spray spread spring spy square squeeze squirrel stable stadium staff stage stairs stamp stand start state stay steak steel stem step stereo stick still sting stock stomach stone stool story stove strategy street strike strong struggle student stuff stumble style subject submit subway success such sudden suffer sugar suggest suit summer sun sunny sunset super supply supreme sure surface surge surprise surround survey suspect sustain swallow swamp swap swarm swear sweet swift swim swing switch sword symbol symptom syrup system table tackle tag tail talent talk tank tape target task taste tattoo taxi teach team tell ten tenant tennis tent term test text thank that theme then theory there they thing this thought three thrive throw thumb thunder ticket tide tiger tilt timber time tiny tip tired tissue title toast tobacco today toddler toe together toilet token tomato tomorrow tone tongue tonight tool tooth top topic topple torch tornado tortoise toss total tourist toward tower town toy track trade traffic tragic train transfer trap trash travel tray treat tree trend trial tribe trick trigger trim trip trophy trouble truck true truly trumpet trust truth try tube tuition tumble tuna tunnel turkey turn turtle twelve twenty twice twin twist two type typical ugly umbrella unable unaware uncle uncover under undo unfair unfold unhappy uniform unique unit universe unknown unlock until unusual unveil update upgrade uphold upon upper upset urban urge usage use used useful useless usual utility vacant vacuum vague valid valley valve van vanish vapor various vast vault vehicle velvet vendor venture venue verb verify version very vessel veteran viable vibrant vicious victory video view village vintage violin virtual virus visa visit visual vital vivid vocal voice void volcano volume vote voyage wage wagon wait walk wall walnut want warfare warm warrior wash wasp waste water wave way wealth weapon wear weasel weather web wedding weekend weird welcome west wet whale what wheat wheel when where whip whisper wide width wife wild will win window wine wing wink winner winter wire wisdom wise wish witness wolf woman wonder wood wool word work world worry worth wrap wreck wrestle wrist write wrong yard year yellow you young youth zebra zero zone zoo`, " ")

func ExtractMacAddress(line string) (string, error) {
	// Attempt to parse the "ff,f1:f5:dd" format
	re1 := regexp.MustCompile(`[,]([a-f0-9:,]+)`)
	match1 := re1.FindStringSubmatch(line)
	if len(match1) > 1 {
		return match1[1], nil
	}

	// Attempt to parse the "1,56:67:ee" format
	re2 := regexp.MustCompile(`1,([a-f0-9:,]+)`)
	match2 := re2.FindStringSubmatch(line)
	if len(match2) > 1 {
		return match2[1], nil
	}

	return "", fmt.Errorf("could not parse hw_address line: %s", line)
}

type DhcpData struct {
	Name       string
	IpAddress  string
	HwAddress  string
	Identifier string
	Lease      string
}

func ParseDhcpLeasesFile(input string) [][]string {
	re := regexp.MustCompile(`{[^{}]*}`)
	blocks := re.FindAllString(input, -1)

	var result [][]string
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		var data []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					data = append(data, strings.TrimSpace(parts[1]))
				}
			}
		}
		result = append(result, data)
	}
	return result
}

func ConvertStringArrayToDhcpDataArray(dataArray [][]string) []DhcpData {
	var data []DhcpData

	for _, entry := range dataArray {
		if len(entry) < 5 {
			//if dhcp lease entry has less than 5 elements skip it
			continue
		}
		hwAddressParts := strings.SplitN(entry[2], ",", 2)
		hAddress := hwAddressParts[1]

		dataItem := DhcpData{
			Name:       entry[0],
			IpAddress:  entry[1],
			HwAddress:  hAddress,
			Identifier: entry[3],
			Lease:      entry[4],
		}
		data = append(data, dataItem)
	}
	return data
}

func MatchHwAddress(data []DhcpData, targetHwAddress string) *DhcpData {
	for i := range data {
		if data[i].HwAddress == targetHwAddress {
			return &data[i]
		}
	}
	return nil
}
