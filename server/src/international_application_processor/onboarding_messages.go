package international_application_processor

import (
	"fmt"
	"html"
	"strings"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
)

func globalCoordinatorSignatureHtml() string {
	return fmt.Sprintf(`
		<p>
			<strong>%v</strong><br/>
			%v<br/>
			Direct Action Everywhere
		</p>`,
		globalCoordinator.Name, globalCoordinator.Role)
}

func networkMemberProgramCoordinatorSignatureHtml() string {
	return fmt.Sprintf(`
		<p>
			<strong>%v</strong><br/>
			%v<br/>
			Direct Action Everywhere
		</p>`,
		networkMemberProgramCoordinator.Name, networkMemberProgramCoordinator.Role)
}

func californiaCoordinatorSignatureHtml() string {
	return fmt.Sprintf(`
		<p>
			<strong>%v</strong><br/>
			%v<br/>
			Direct Action Everywhere<br/>
			%v
		</p>
	`, californiaCoordinator.Name, californiaCoordinator.Role, californiaCoordinator.Pronouns)
}

func (b *onboardingEmailMessageBuilder) nearSFBayChapter() (mailer.Message, error) {
	var msg mailer.Message
	msg.FromName = sfBayCoordinator.Name
	msg.FromAddress = sfBayCoordinator.Address
	msg.ToName = b.fullName
	msg.ToAddress = b.email
	msg.Subject = "Join your local Direct Action Everywhere chapter!"

	var body strings.Builder

	fmt.Fprintf(&body, "<p>Hey %v!</p>", html.EscapeString(b.firstName))

	body.WriteString(
		`<p>
			Thank you for reaching out about getting involved with Direct Action
			Everywhere.
		</p>
		<p>
			Great news—you're near the SF Bay Area chapter.
			We’d love to see you at an upcoming event!
		</p>
		<p>
			Here are some ways to get started:
			<ul>
		`)

	nextEventLink := getFacebookEventLinkOrEmptyString(b.nextEvent)
	if len(nextEventLink) > 0 {
		fmt.Fprintf(
			&body,
			"<li>Come to our next event: %v</li>",
			nextEventLink)
	}

	body.WriteString(`
				<li>Find all of our events at <a href="https://dxe.io/events">dxe.io/events</a>, including community gatherings, trainings, and actions.</li>
				<li>Become a member at <a href="https://dxe.io/apply">dxe.io/apply</a>. When you’re ready to dive deeper, fill out this quick application.</li>
				<li>Sign the <a href="https://righttorescue.com/">letter to support the right to rescue</a>.</li>
			</ul>
		</p>
		<p>Feel free to reply with any questions—I’m happy to help.</p>
		<p>Hope to see you soon!</p>
	`)

	fmt.Fprintf(&body,
		`<p><strong>%v</strong><br/>
		DxE Organizer</p>`,
		sfBayCoordinator.Name)

	msg.BodyHTML = body.String()

	return msg, nil
}

func (b *onboardingEmailMessageBuilder) nearNonSFBayChapter() (mailer.Message, error) {
	var msg mailer.Message
	msg.FromName = globalCoordinator.Name
	msg.FromAddress = globalCoordinator.Address
	msg.ToName = b.fullName
	msg.ToAddress = b.email
	msg.CC = getChapterEmailsWithFallback(b.chapter, getChapterEmailFallback(b.state))
	msg.Subject = "Join your local Direct Action Everywhere chapter!"

	var body strings.Builder

	fmt.Fprintf(&body, "<p>Hey %v!</p>", html.EscapeString(b.firstName))

	body.WriteString(`
		<p>
			I wanted to reach out about your inquiry of getting involved with
			DxE’s international network.
		</p>
		<p>
			Great news, there is a DxE chapter near you! I’ve CC’d the
			organizers on this email so you can be in contact, and I’ve included
			the chapter’s information below so you can reach out, get involved,
			and start taking action with them!
		</p>
	`)

	body.WriteString("<p>")
	if b.chapter.FbURL != "" {
		fmt.Fprintf(&body, `<a href="%v">DxE %v Facebook page</a><br />`,
			b.chapter.FbURL, b.chapter.Name)
	}
	if b.chapter.InstaURL != "" {
		fmt.Fprintf(&body, `<a href="%v">DxE %v Instagram</a><br />`,
			b.chapter.InstaURL, b.chapter.Name)
	}
	if b.chapter.TwitterURL != "" {
		fmt.Fprintf(&body, `<a href="%v">DxE %v Twitter</a><br />`,
			b.chapter.TwitterURL, b.chapter.Name)
	}
	if b.chapter.Email != "" {
		fmt.Fprintf(&body, `<a href="mailto:%v">DxE %v Email</a><br />`,
			b.chapter.Email, b.chapter.Name)
	}
	body.WriteString("</p>")

	nextEventLink := getFacebookEventLinkOrEmptyString(b.nextEvent)
	if len(nextEventLink) > 0 {
		fmt.Fprintf(
			&body,
			"<p>You can also find details of our next event here: %v</p>",
			nextEventLink)
	}

	body.WriteString(`
		<p>
			In the meantime you can
			<a href="https://righttorescue.com/">sign a letter to support the
			right to rescue</a>.
		</p>
		<p>
			Let me know if you have any questions or if you aren’t able to
			connect with your local chapter.
		</p>
		<p>Hope that you can join us!</p>
		`)

	body.WriteString(globalCoordinatorSignatureHtml())

	msg.BodyHTML = body.String()

	return msg, nil
}

func (b *onboardingEmailMessageBuilder) caOrganizerNotNearAnyChapter() (mailer.Message, error) {
	var msg mailer.Message
	msg.FromName = californiaCoordinator.Name
	msg.FromAddress = californiaCoordinator.Address
	msg.ToName = b.fullName
	msg.ToAddress = b.email

	msg.Subject = "Getting involved with Direct Action Everywhere"

	var body strings.Builder

	fmt.Fprintf(&body, `<p>Hi %v,</p>`, html.EscapeString(b.firstName))

	body.WriteString(`
		<p>
			Thank you for your interest in joining the DxE Network.
			Right now, DxE is actively seeking new organizers and chapters in
			California.
			There is no active chapter in your area and I am excited to help you
			launch one!
		</p>`)
	fmt.Fprintf(&body, `
		<p>
			To begin, please find 1-2 other people in your area who might be
			interested in helping you organize a chapter. Once you have
			identified them, reach out to me by email at %v to schedule a call.
			Don’t hesitate to email me with any questions in the meantime.
		</p>
		<p>
			I’m looking forward to hearing back from you,
		</p>
		`, californiaCoordinator.Address)

	body.WriteString(californiaCoordinatorSignatureHtml())

	msg.BodyHTML = body.String()

	return msg, nil
}

func (b *onboardingEmailMessageBuilder) nonCaOrganizerNotNearAnyChapter() (mailer.Message, error) {
	var msg mailer.Message
	msg.FromName = globalCoordinator.Name
	msg.FromAddress = globalCoordinator.Address
	msg.ToName = b.fullName
	msg.ToAddress = b.email
	msg.Subject = "Getting involved with Direct Action Everywhere"

	var body strings.Builder
	fmt.Fprintf(&body, `<p>Hi %v,</p>`, html.EscapeString(b.firstName))

	body.WriteString(`
		<p>
			Thank you for your interest in joining the DxE Network.
			I am part of the international coordination (IC) team and I am here
			to help you start a DxE chapter in your area. Our onboarding process
			involves four key steps:
		</p>
		<ol>
			<li>
				Finding 1-2 other people in your area who are interested in
				helping start the chapter, and then setting up a call with a
				member of the IC team. During this call, we will explain the
				whole onboarding process and you will also be assigned a mentor.
			</li>
			<li>
				Completing 5 short training sessions that cover important
				information about DxE and how to organize your first events.
			</li>
			<li>
				Organizing your first action and community event. Don’t worry,
				you will have a lot of time and support to make this happen!
			</li>
			<li>
				Debriefing your action with your mentor and completing the
				final onboarding steps for you and your chapter. At that point,
				you will be an official DxE organizer in an official DxE
				chapter!
			</li>
		</ol>

		<p>
			To begin, please find 1-2 other people in your area who are
			interested in helping you start a DxE chapter. Once you have
			identified them, reach out to me by email to schedule our first
			call.
			Don’t hesitate to email me with any questions.
		</p>
		<p>
			I’m looking forward to hearing back from you!
		</p>
	`)

	body.WriteString(globalCoordinatorSignatureHtml())

	msg.BodyHTML = body.String()

	return msg, nil
}

func (b *onboardingEmailMessageBuilder) participantNotNearAnyChapter() (mailer.Message, error) {
	var msg mailer.Message
	msg.FromName = networkMemberProgramCoordinator.Name
	msg.FromAddress = networkMemberProgramCoordinator.Address
	msg.ToName = b.fullName
	msg.ToAddress = b.email
	msg.Subject = "Getting involved with Direct Action Everywhere"

	var body strings.Builder
	fmt.Fprintf(&body, `<p>Hi %v,</p>`, html.EscapeString(b.firstName))

	body.WriteString(`
		<p>
			Thank you for your interest in joining the DxE Network.
			We recently launched the Network Member Program which is a platform
			for people who are interested in taking action with DxE but don’t
			have a chapter near them or have the capacity to organize a chapter
			themselves.
			The only steps to complete to officially become a member are:
		</p>
		<ol>
			<li><a href="https://dxe.io/networkmember">Watch a video</a></li>
			<li><a href="https://dxe.io/networkmemberquiz">Complete a short quiz</a></li>
		</ol>
		<p>
			As a Network Member you’ll be able to connect with other activists
			around the US and the world, join DxE’s campaigns, receive mentoring
			to develop your activism skills, and become eligible for financial
			support to attend events and actions in person.
		</p>
		<p>
			If you have any questions or would like help getting started, just
			reply to this email—I’m happy to assist. I hope to welcome you to
			the Network soon!
		</p>
	`)

	body.WriteString(networkMemberProgramCoordinatorSignatureHtml())

	msg.BodyHTML = body.String()

	return msg, nil
}

func getFacebookEventLinkOrEmptyString(event *model.ExternalEvent) string {
	if event == nil || len(event.ID) == 0 {
		return ""
	}

	return fmt.Sprintf(`<a href="https://facebook.com/events/%v">%v</a>`, event.ID, event.Name)
}
