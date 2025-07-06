import {
  ConsoleLogger,
  DeviceManagementKitBuilder,
} from "@ledgerhq/device-management-kit";
import { webHidTransportFactory } from "@ledgerhq/device-transport-kit-web-hid";
 
// Initialize the DMK with a console logger and WebHID transport
/*
export const dmk = new DeviceManagementKitBuilder()
  .addLogger(new ConsoleLogger())
  .addTransport(webHidTransportFactory) // Transport is required!
  .build();
*/
export const dmk = new DeviceManagementKitBuilder()
  .addTransport(webHidTransportFactory) // Transport is required!
  .build();


