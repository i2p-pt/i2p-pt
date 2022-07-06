package common

import (
	"log"
	"os"

	"github.com/eyedeekay/i2pkeys"
	"github.com/eyedeekay/sam3"
)

func I2PSession(name, samaddr, keyspath string) (*sam3.StreamSession, error) {
	log.Printf("Starting and registering I2P session...")
	sam, err := sam3.NewSAM(sam3.SAMDefaultAddr(samaddr))
	if err != nil {
		log.Fatalf("error connecting to SAM to %s: %s", sam3.SAMDefaultAddr(samaddr), err)
	}
	keys, err := setupkeys(keyspath, sam)
	if err != nil {
		return nil, err
	}
	stream, err := sam.NewStreamSession(name, *keys, sam3.Options_Wide)
	return stream, err
}

func setupkeys(keyspath string, sam *sam3.SAM) (keys *i2pkeys.I2PKeys, err error) {
	if sam == nil {
		sam, err = sam3.NewSAM(sam3.SAMDefaultAddr("127.0.0.1:7656"))
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(keyspath + ".i2p.private"); os.IsNotExist(err) {
		f, err := os.Create(keyspath + ".i2p.private")
		if err != nil {
			log.Fatalf("unable to open I2P keyfile for writing: %s", err)
		}
		defer f.Close()
		tkeys, err := sam.NewKeys()
		if err != nil {
			log.Fatalf("unable to generate I2P Keys, %s", err)
		}
		keys = &tkeys
		err = i2pkeys.StoreKeysIncompat(*keys, f)
		if err != nil {
			log.Fatalf("unable to save newly generated I2P Keys, %s", err)
		}
	} else {
		tkeys, err := i2pkeys.LoadKeys(keyspath + ".i2p.private")
		if err != nil {
			log.Fatalf("unable to load I2P Keys: %e", err)
		}
		keys = &tkeys
	}
	return keys, nil
}
