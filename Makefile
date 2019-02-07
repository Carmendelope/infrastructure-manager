#
#  Copyright 2018 Nalej
# 

# Name of the target applications to be built
APPS=infrastructure-manager

# Use global Makefile for common targets
export
%:
	$(MAKE) -f Makefile.golang $@
