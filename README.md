# Overview

gounit augments the go testing framework by features for test-driven
development the way I understand it:

* Package-, file-, suite and test-names outline the production
  code's documentation. 

* Tests are implemented from the most simple system-property tests ---
  getting us closer to the overall goal --- to the most intricate
  properties.  I.e. the order of tests matters for understanding the
  (tested) implementation.

* Test-implementations specify the production code's features, i.e.
  document its API.

* Tests should run fast, i.e. as the developer I shouldn't need to wait
  for the tests to run through.

To achieve the above it is useful to have systematic fine grained
control over the "noise" in test-code and its concurrent execution.
gounit provides for this purpose mainly the types gounit.Suite and
gounit.T.  See the following animation of a gounit prototype
implementation to get a first working-impression.

![simple gounit use-case](gounit.gif)

