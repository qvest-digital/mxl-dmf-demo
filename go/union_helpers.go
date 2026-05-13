package mxl

/*
#include <mxl/flowinfo.h>

// Cgo cannot address members of anonymous C unions directly. These trivial
// helpers return typed pointers into the discrete/continuous arms of
// mxlFlowConfigInfo's union. The Go caller must check the format field
// before dereferencing.

static inline mxlDiscreteFlowConfigInfo const* mxl_go_config_discrete(mxlFlowConfigInfo const* c) {
    return &c->discrete;
}

static inline mxlContinuousFlowConfigInfo const* mxl_go_config_continuous(mxlFlowConfigInfo const* c) {
    return &c->continuous;
}
*/
import "C"

import "unsafe"

func discreteFromUnion(c *C.mxlFlowConfigInfo) unsafe.Pointer {
	return unsafe.Pointer(C.mxl_go_config_discrete(c))
}

func continuousFromUnion(c *C.mxlFlowConfigInfo) unsafe.Pointer {
	return unsafe.Pointer(C.mxl_go_config_continuous(c))
}
