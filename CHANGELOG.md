
v4.0.0 / 2021-07-22
===================

  * build: fix semantic import versioning (#63)

v3.1.0 / 2021-07-16
===================

  * build: update golang dependencies to modern versions (#60)
  * build: use github actions for CI (#61)
  * build: remove vendored code (#62)
  * tests: fixes for 32-bit platforms related to integers (#57)
  * Update README with modern PCP and grafana-pcp (over Vector)

v3.0.1 / 2017-10-30
===================

  * metrics: use a global InstanceDomain for PCPHistogram, fixes #54 (#55)
  * build: add golangci-lint and run lint on CI again (#56) 

v3.0.0 / 2017-10-30
===================

  * metrics: add support for multidimensional composite metrics (#50)
  * speed: remove logging from library (#49) (**BREAKING**)
  * add mmvdump tests (#48)

v2.0.0 / 2017-04-05
===================

  * examples: use log.Fatal instead of panic in all examples (#45)
  * examples: add example which exposes Go runtime metrics (#44)
  * Replace logrus with zap (#43)
  * Port Speed to Windows (#41)

v1.0.0 / 2016-10-28
===================
