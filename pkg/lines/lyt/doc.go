// Package lyt provides the means to layout components in a given area.
// It provides for this purpose the Stacker, Chainer and Dimer
// interfaces which need to be implemented by layouted components.
// Whereas a layouted component MUST only implement EITHER a Stacker OR
// a Chainer interface NOT BOTH but can implement none of them.  All
// components must implement the Dimer interface. E.g.
//
//     +--------------------------+
//     |            1             |
//     |                          |
//     +-------+---------+--------+
//     |   2   |    3    |   4    |
//     |       |         |        |
//     +-------+---------+--------+
//     |                          |
//     |            5             |
//     +--------------------------+
//
// Can be realized as a Stacker which provides the components (1),
// (2,3,4) and (5) whereas (2,3,4) is implemented as a Chainer.  Use an
// instance of Manger with a set Root Dimer to calculate a layout:
//
//     m := Manager{ Root: myDimer }
//     m.Reflow()
//
// The second line calculates the layout of the Root-Dimer.  The root
// Dimer's size defines the area for which the layout is calculated.  Is
// set Root implementing either the Stacker or Chainer interface the
// layout of provided Dimers by this implementation is calculated as
// well.  If one of these provided Dimers implements either of those
// interfaces its provided Dimers' layout is also calculated and so on.
// Note a manager doesn't hold any data besides the root component.
// I.e. it can not notice if a new component was added or a layouted
// component whishes to change its size.  For the former call:
//
//     m.Reflow(newDimer)
//
// for the later
//
//     m.UpdateHeight(dimerToUpdate, height)
//
// or
//
//     m.UpdateWidth(dimerToUpdate, width)
//
// Overflowing components of a layout are clipped accordingly
// underflowing components receive margins to fill the unused space.
// Use the methods Area, Clip and Margin to evaluate the available/used
// area of a layouted Dimer.  See the Dim-constructors DimFilling,
// DimFillingWidth, DimFillingHeight and DimFixed for how the layout
// calculation can be influenced by providing different kinds of
// Dim-instances.  Using only these Dim-constructors for
// Dimer-implementations makes sure that the layout manager has all
// needed information to calculate a layout available.
package lyt
