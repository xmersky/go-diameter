// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Converts Wireshark diameter dictionaries to go-diameter format.
// Use: wireshark-dict-tool < wireshark-dict.xml > new-dict.xml
//
// Some wireshark dictionaries must be slightly fixed before they can
// be converted by this tool.

package main

// TODO: Improve the parser and fix AVP properties during conversion:
// <avp name=".." code=".." must="" may="" must-not="" may-encrypt="">

import (
	"encoding/xml"
	"log"
	"os"

	"github.com/fiorix/go-diameter/diam/diamdict"
)

func main() {
	wsd, err := Load(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	var new_dict = &diamdict.File{}
	for _, app := range wsd.App {
		new_app := &diamdict.App{
			Id:   app.Id,
			Type: app.Type,
			Name: app.Name,
		}
		copy_vendors(wsd.Vendor, new_app)
		copy_commands(app.Cmd, new_app)
		copy_avps(app.AVP, new_app)
		new_dict.App = append(new_dict.App, new_app)
	}
	os.Stdout.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"))
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "\t")
	enc.Encode(new_dict)
}

func copy_vendors(src []*Vendor, dst *diamdict.App) {
	for _, vendor := range src {
		dst.Vendor = append(dst.Vendor, &diamdict.Vendor{
			Id:   vendor.Id,
			Name: vendor.Name,
		})
	}
}

func copy_commands(src []*Cmd, dst *diamdict.App) {
	for _, cmd := range src {
		new_cmd := &diamdict.CMD{
			Code:  cmd.Code,
			Name:  cmd.Name,
			Short: cmd.Name,
		}
		copy_cmd_rules(cmd.Request.Fixed.Rule, &new_cmd.Request, false)
		copy_cmd_rules(cmd.Request.Required.Rule, &new_cmd.Request, true)
		copy_cmd_rules(cmd.Request.Optional.Rule, &new_cmd.Request, false)
		copy_cmd_rules(cmd.Answer.Fixed.Rule, &new_cmd.Answer, false)
		copy_cmd_rules(cmd.Answer.Required.Rule, &new_cmd.Answer, true)
		copy_cmd_rules(cmd.Answer.Optional.Rule, &new_cmd.Answer, false)
		dst.CMD = append(dst.CMD, new_cmd)
	}
}

func copy_cmd_rules(src []*Rule, dst *diamdict.CMDRule, required bool) {
	for _, req := range src {
		dst.Rule = append(dst.Rule, &diamdict.Rule{
			AVP:      req.Name,
			Required: required,
			Min:      req.Min,
			Max:      req.Max,
		})
	}
}

func copy_avps(src []*AVP, dst *diamdict.App) {
	for _, avp := range src {
		new_avp := &diamdict.AVP{
			Name: avp.Name,
			Code: avp.Code,
		}
		if avp.Type.Name == "" && avp.Grouped != nil {
			new_avp.Data = diamdict.Data{TypeName: "Grouped"}
		} else {
			new_avp.Data = diamdict.Data{TypeName: avp.Type.Name}
		}
		switch avp.MayEncrypt {
		case "yes":
			new_avp.MayEncrypt = "Y"
		case "no":
			new_avp.MayEncrypt = "N"
		default:
			new_avp.MayEncrypt = "-"
		}
		switch avp.Mandatory {
		case "must":
			new_avp.Must = "M"
		case "may":
			new_avp.May = "P"
		default:
			new_avp.Must = ""
		}
		if new_avp.May != "" {
			switch avp.Protected {
			case "may":
				new_avp.May = "P"
			default:
				new_avp.May = ""
			}
		}
		for _, p := range avp.Enum {
			new_avp.Data.Enum = append(new_avp.Data.Enum,
				&diamdict.Enum{
					Name: p.Name,
					Code: p.Code,
				})
		}
		for _, grp := range avp.Grouped {
			for _, p := range grp.GAVP {
				new_avp.Data.Rule = append(new_avp.Data.Rule,
					&diamdict.Rule{
						AVP: p.Name,
						Min: p.Min,
						Max: p.Max,
					})
			}
			for _, p := range grp.Required.Rule {
				new_avp.Data.Rule = append(new_avp.Data.Rule,
					&diamdict.Rule{
						AVP:      p.Name,
						Required: true,
						Min:      p.Min,
						Max:      p.Max,
					})
			}
			for _, p := range grp.Optional.Rule {
				new_avp.Data.Rule = append(new_avp.Data.Rule,
					&diamdict.Rule{
						AVP:      p.Name,
						Required: false,
						Min:      p.Min,
						Max:      p.Max,
					})
			}
		}
		dst.AVP = append(dst.AVP, new_avp)
	}
}
