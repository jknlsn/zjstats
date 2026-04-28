package collector

// #cgo LDFLAGS: -framework IOKit -framework CoreFoundation
// #include <IOKit/IOKitLib.h>
// #include <CoreFoundation/CoreFoundation.h>
//
// int gpuUtilization() {
//     int utilization = -1;
//     io_iterator_t iter = 0;
//     kern_return_t kr = IOServiceGetMatchingServices(0, IOServiceMatching("AGXAccelerator"), &iter);
//     if (kr != KERN_SUCCESS || iter == 0) {
//         if (iter) IOObjectRelease(iter);
//         return -1;
//     }
//
//     io_service_t service = IOIteratorNext(iter);
//     if (service) {
//         // "Device Utilization %" is nested inside "PerformanceStatistics" dictionary.
//         CFStringRef statsKey = CFStringCreateWithCString(kCFAllocatorDefault, "PerformanceStatistics", kCFStringEncodingUTF8);
//         CFTypeRef statsProp = IORegistryEntryCreateCFProperty(service, statsKey, kCFAllocatorDefault, 0);
//         if (statsProp) {
//             if (CFGetTypeID(statsProp) == CFDictionaryGetTypeID()) {
//                 CFDictionaryRef stats = (CFDictionaryRef)statsProp;
//                 CFStringRef utilKey = CFStringCreateWithCString(kCFAllocatorDefault, "Device Utilization %", kCFStringEncodingUTF8);
//                 CFTypeRef utilProp = CFDictionaryGetValue(stats, utilKey);
//                 if (utilProp && CFGetTypeID(utilProp) == CFNumberGetTypeID()) {
//                     CFNumberGetValue((CFNumberRef)utilProp, kCFNumberIntType, &utilization);
//                 }
//                 if (utilKey) CFRelease(utilKey);
//             }
//             CFRelease(statsProp);
//         }
//         if (statsKey) CFRelease(statsKey);
//         IOObjectRelease(service);
//     }
//     IOObjectRelease(iter);
//     return utilization;
// }
import "C"

func collectGPU() int {
	return int(C.gpuUtilization())
}
