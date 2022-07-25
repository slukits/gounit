// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides an unopinionated, well tested and documented,
// terminal UI library which does the heavy lifting for you when it
// comes to
//
//     - an architecture that is robust against race conditions
//     - event handling
//     - feature handling
//     - layout handling
//
// The motivation is to provide an UI-library with few powerful features
// that let its user quickly implement an terminal ui exactly the way it
// is desired.  Everything should be at service if requested but not in
// the way if not.  If for example no message-bar used there is none.
// If the message bar is used it is made available with reasonable
// defaults.  If these defaults are unreasonable for the use case at
// hand they can be changed...
//
// lines wraps the package https://github.com/gdamore/tcell which does
// the heavy lifting on the terminal side.  I didn't make the effort to
// wrap the constants and types which are defined by tcell and are used
// for event-handling and styling.  I.e. you will have to make yourself
// acquainted with tcell's Key constants its ModeMap constants, its
// AttrMask constants, its Style type and Color handling as needed.
//
// Events and Concurrency
//
// lines is built around event handling.
//
//     import "github.com/slukits/lines"
//
//     ee, err := lines.New(nil)
//     if err != nil {
//         fmt.Fatalf("can't acquire events: %v", err)
//     }
//     ee.Listen()
//
// New provides you with an Events-instance which starts reporting
// events when Listen is called.  Listen will block until 'q', ctrl-c or
// ctrl-d is pressed or ee.QuitListening() is called.
//
// We can access lines features by implementing UI-components.
//
//     type Global struct { lines.UIComponent }
//
//     func (c *Global) OnInit(e *lines.Env) {
//         fmt.Fprintf(e.MessageBar, "%s %s", "hello", "world")
//     }
//
//     func (c *Global) OnQuit() {
//         fmt.Println("May you be fearless, joyful and compassionate.")
//     }
//
// Initializing our Events-instance with lines.New(&Global{}) will
// register our first component and report the Init-event to it.  The
// Init-event is always the first event that is reported to a component
// and it is reported exactly once.
//
// Except for the quit-listener an event listener is always provided
// with an environment instance (lines.Env-instance).  An environment is
// always bound to the component on which the event was reported and its
// screen area.  It provides information about the event and means to
// communicate back to the reporting Events instance. E.g.
//
//     func (c *Global) OnKey(e *lines.Env, tcell.Key, tcell.ModMask) {
//         fmt.Fprintf(e, "'%s'-pressed", e.Evt.(tcell.EventKey).Name())
//         e.StopBubbling()
//     }
//
// Is an environment printed to after the listener it was provided to
// has returned it will panic.
//
//     func (c *Global) OnKey(e *lines.Env, tcell.Key, tcell.ModMask) {
//         go func() {
//             time.Sleep(1*time.Second)
//             fmt.Fprint(e.Statusbar, "awoken") // will panic
//         }()
//     }
//
// but what we can do instead is
//
//     func (c *Global) OnKey(e *lines.Env, tcell.Key, tcell.ModMask) {
//         go func(ee *lines.Events) {
//             time.Sleep(1*time.Second)
//             ee.UpdateComponent(c, nil, func(e *lines.Env) {
//                  fmt.Fprint(e.Statusbar, "awoken") // will not panic
//             })
//         }(e.EE)
//     }
//
// this little hoop we have to jump through gives us concurrency save
// access to UI-components leveraging the event queue.
//
// Layout
//
// To get a more sophisticated screen-layout than message bar, statusbar
// and "the rest of the screen" we need to implement more ui-components.
//
//     type Cmp1 struct { lines.UIComponent }
//
//     func (c *Cmp1) Init(e *lines.Evn) {
//         fmt.Fprint(e.Fmt(lines.CenterBoth), "c1")
//     }
//
//     type Cmp2 struct { lines.UIComponent }
//
//     func (c *Cmp2) Init(e *lines.Evn) {
//         fmt.Fprint(e.Fmt(lines.CenterBoth), "c2")
//     }
//
//     type Cmp3 struct { lines.UIComponent }
//
//     func (c *Cmp3) Init(e *lines.Evn) {
//         fmt.Fprint(e.Fmt(lines.CenterBoth), "c3")
//     }
//
//     type Cmp4 struct { lines.UIComponent }
//
//     func (c *Cmp4) Init(e *lines.Evn) {
//         fmt.Fprint(e.Fmt(lines.CenterBoth), "c4")
//     }
//
// Then we add our ui-components to the layout of the global environment
//
//     func (c *Global) Init(e *lines.Env) {
//         c2 := &Cmp2{}
//         e.Before(nil, &Cmp1{}, c2, &Cmp3{})
//         e.NextTo(c2, &Cmp4{})
//     }
//
// e.Before and e.NextTo adds our components to the layout.  Is the
// first argument of Before nil then the component is added at the end
// of the ui-stack given environment is associated with.  NextTo on the
// other hand does not stack but chain horizontally.  Hence we will get
// below screen layout
//
//     +-----------------------------------------+
//     |                                         |
//     |                   c1                    |
//     |                                         |
//     +-------------------+---------------------+
//     |                   |                     |
//     |       c2          |          c4         |
//     |                   |                     |
//     +-------------------+---------------------+
//     |                                         |
//     |                   c3                    |
//     |                                         |
//     +-----------------------------------------+
//
// No worries the borders I drew above are not on the screen by default.
// Embedding *lines.UIComponent* provides us with an implementation of
// the Dimer-interface, i.e. the Dim-method.  Adding to the Init-method
// of Cmp1
//
//     c1.Dim().SetHeight(1)
//
// and to Cmp3's Init-method the line
//
//     c3.Dim().SetHeight(1)
//
// will result in the layout
//
//     +-----------------------------------------+
//     |                   c1                    |
//     +-------------------+---------------------+
//     |                   |                     |
//     |                   |                     |
//     |                   |                     |
//     |       c2          |          c4         |
//     |                   |                     |
//     |                   |                     |
//     |                   |                     |
//     +-------------------+---------------------+
//     |                   c3                    |
//     +-----------------------------------------+
//
// Lets assume we are listening for rune-input at the second component.
//
//     func (c *Cmp2) OnRune(e *lines.Env, r rune) {
//         fmt.Fprintf(e.MessageBar, "received rune '%c'", r)
//     }
//
// Note that the environment we get in the above Rune-listener will
// print automatically only to c2's screen-area.  Note also that you can
// resize your terminal-window and the proportions and alignment of the
// content is automatically updated.
//
// The moment we are listening for an event on a component the component
// becomes selectable, i.e. if the user presses (often enough) the Tab
// key eventually our component will get the focus and can receive
// events.  By default always the last component added to the layout has
// the focus and with e.MoveFocus(cmp) we can ask for a focus change
// after the listener has returned.
//
// Please see the ``numpad'' example in the examples directory to see a
// full application which uses all the core principles of this package.
//
// Enjoy!
package lines
