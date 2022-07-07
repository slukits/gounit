# Overview

gounit augments the go testing framework by features for test-driven
development the way I understand it:

    * Package-, file-, suite and test-names outline the production
      code's documentation. 

    * Test-implementations specify the production code's features
      and document its API.

To achieve the later it is useful to have systematic fine grained
control over the "noise" in test-code.  gounit provides for this purpose
mainly the types gounit.Suite and gounit.T.  See the following animation
of a gounit prototype implementation to get a first working-impression.

![simple gounit use-case](gounit.gif)