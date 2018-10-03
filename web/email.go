package main

/**
 * Copyright (C) 2018 Preetam Jinka
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"bytes"
	"fmt"
	"html/template"
	textTemplatePkg "html/template"
)

var textTemplate = textTemplatePkg.Must(textTemplatePkg.New("text").Parse(`Hi there!

{{ .Content }}

Cheers!
--Preetam

Copyright (c) 2018 Preetam Jinka
You are receiving this email because you signed up for Transverse.
www.transverseapp.com
`))

var emailTemplate = template.Must(template.New("email").Parse(`
<!DOCTYPE html>
<html>

<body>
<table style="font-family: Verdana, sans-serif; font-color: #292b2c; font-size: 14px; background-color: white; margin: 0 auto; min-width: 70%; max-width: 90%;">
  <tr><td style="padding: 30px">
	<div>
	  <a href="https://www.transverseapp.com/"><img src="https://www.transverseapp.com/img/email-logo.png" width=125 alt="Transverse"/></a>
	</div>

	<div>
		<p>Hi there!</p>
		{{ .HTMLContent }}
		<p>Cheers!<br>&mdash;Preetam</p>
	</div>

	<div style="color: #888; font-size: 12px; text-align: center; padding-top: 10px;">
	  <p>Copyright &copy; 2018 Preetam Jinka</p>
	  <p>You are receiving this email because you signed up for Transverse.</p>
	  <a style="color: #888" href="https://www.transverseapp.com/">transverseapp.com</a>
	</div>
  </td></tr>
</table>
</body>

</html>
`))

func CodeEmail(codeType, code string) string {
	return fmt.Sprintf(`
<p>Your %s code is %s.<p>
<p>This is only valid for a few minutes, so you better hurry!</p>
`, codeType, code)
}

func SendEmail(to, subject, content, htmlContent string) error {
	if htmlContent != "" {
		buf := &bytes.Buffer{}
		err := emailTemplate.Execute(buf, map[string]interface{}{
			"HTMLContent": template.HTML(htmlContent),
		})
		if err != nil {
			return err
		}
		htmlContent = string(buf.Bytes())
	}

	buf := &bytes.Buffer{}
	err := textTemplate.Execute(buf, map[string]interface{}{
		"Content": content,
	})
	if err != nil {
		return err
	}
	content = string(buf.Bytes())

	if mg.ApiKey() == "" {
		fmt.Println("Email")
		fmt.Println("===")
		fmt.Println("To:", to)
		fmt.Println("Subject:", subject)
		fmt.Println("Content:", content)
		if htmlContent != "" {
			fmt.Println("HTML Content:", htmlContent)
		}
		return nil
	}

	msg := mg.NewMessage("Transverse <noreply@transverseapp.com>", subject, content, to)
	msg.SetReplyTo("Preetam <transverse@preet.am>")
	msg.SetHtml(htmlContent)

	_, _, err = mg.Send(msg)
	return err
}
