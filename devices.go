package main

import (
	"fmt"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"time"
)

var logger = shim.NewLogger("DIChaincode")

type IMEI_Holder struct {
	IMEIs 	[]string `json:"imeis"`
}

type Device struct {
	DeviceName     string `json:"devicename"`
	DeviceModel    string `json:"devicemodel"`
	DateOfManf     string `json:"dateofmanf"`
	ConsignmentNumber string `json:"consignmentnumber"`
	DateOfDelivery string `json:"dateofdelivery"`
	DateOfReceipt  string `json:"dateofreceipt"`
	DateOfSale     string `json:"dateofsale"`
	OldIMEI        string `json:"oldimei"`
	IMEI	       string `json:"imei"`
	Status         string `json:"status"`
	SoldBy         string `json:"soldby"`
	Owner          string `json:"owner"`
}

type SimpleChainCode struct {
}

func (t *SimpleChainCode) Init(stub shim.ChaincodeStubInterface, function string, args[] string) ([]byte, error ) {
	
	var imeiIds IMEI_Holder
	
	bytes, err := json.Marshal(imeiIds);
	
	if err != nil { return nil, errors.New("Error creating IMEI_Holder record") }
	
	err = stub.PutState("imeiIds", bytes)
	
	return nil, err
	 
} 

func (t *SimpleChainCode) Invoke(stub shim.ChaincodeStubInterface, function string, args[] string) ([]byte, error) {
	
	if function == "create_device" && len(args) == 1 {
		return	t.createDevice(stub, args[0])
	} else if function == "create_device" && len(args) > 1 {
		if len(args) != 4 {fmt.Printf("Incorrect input data passed. Cannot process creation"); return nil, errors.New("Invalid input arguments for device creation")} 
		return	t.createDeviceUsingForm(stub, args)
	} else {
		d, err := t.get_device(stub, args[0])
		
		if err != nil { fmt.Printf("error retrieving device details"); return nil, errors.New("error retrieving device details")}
		
		if function == "TRF_TO_WH" { return t.tranfer_to_WareHouse(stub, d, "VENDOR", args[1], args[2], "WAREHOUSE")
		} else if function == "ACPT_FROM_VENDOR" { return t.accept_from_vendor(stub, d, "WAREHOUSE", args[1], "WAREHOUSE")
		} else if function == "TRF_TO_STRE" { return t.tranfer_to_store(stub, d, "WAREHOUSE", args[1], args[2], "STORE")
		} else if function == "ACPT_FROM_WAREHOUSE" { return t.accept_from_warehouse(stub, d, "STORE", args[1], "STORE")	
		} else if function == "TRF_TO_CUST" { return t.tranfer_to_customer(stub, d, "STORE", args[1], args[2], "STORE")
		} else if function == "RTN_FROM_CUST" { return t.return_from_customer(stub, d, "STORE", args[1], "STORE")					
		} else if function == "EXCHANGE_DEV" { 
			oldDev, err := t.get_device(stub, args[2])
			if err != nil {fmt.Printf("unable to get old device"); return nil, errors.New("Unable to return old device")}
			return t.exchange_device(stub, oldDev, d, "STORE", args[1], "STORE")
		} else if function == "RTN_TO_WAREHOUSE" { return t.return_to_warehouse(stub, d, "STORE", args[1], args[2], "WAREHOUSE")
		} else if function == "ACPT_FROM_STRE" { return t.return_from_store(stub, d, "WAREHOUSE", args[1], "WAREHOUSE")		
		} else if function == "RTN_TO_VENDOR" { return t.return_to_vendor(stub, d, "WAREHOUSE", args[1], args[2], "VENDOR")
		} else if function == "ACPT_RTN_FROM_WAREHOUSE" { return t.return_from_warehouse(stub, d, "VENDOR", args[1], "VENDOR")		
		} 
	}		
	return nil, nil
}

func (t *SimpleChainCode) Query(stub shim.ChaincodeStubInterface, function string, args[] string) ([]byte, error) {
	
	if function == "get_device_details" {
		d, err := t.get_device(stub, args[0])
		if err != nil { fmt.Printf("error retrieving device details"); return nil, errors.New("error retrieving device details")}
		return t.get_dev_details(stub, d)
	}  else if function == "check_unique_imei" {
		return t.check_unique_imei(stub, args[0])
	} else if function == "get_devices" {
		return t.get_devices(stub)
	}
	return nil, nil
}

func (t *SimpleChainCode) createDevice(stub shim.ChaincodeStubInterface, imeiId string) ([]byte, error) {
	
	var d Device
	var err error
	var IMEI_Ids IMEI_Holder
	
	DeviceName  := "\"deviceName\":\"LENOVO\", "
	DeviceModel := "\"devicemodel\":\"VIBE\", "
	DateOfManf  := "\"dateofmanf\":\"''03-12-2016''\" , "
	DateOfSale  := "\"dateofsale\":\"UNDEFINED\", "
	OldIMEI     := "\"oldimei\":\"UNDEFINED\", "
	IMEI_ID     := "\"imei\":\""+imeiId+"\", "
	Status     	:= "\"status\":\"CREATED\", "
	SoldBy     	:= "\"soldby\":\"UNDEFINED\", "
	Owner     	:= "\"owner\":\"VENDOR\" "
	
	json_device := " {" +DeviceName+DeviceModel+DateOfManf+DateOfSale+OldIMEI+IMEI_ID+Status+SoldBy+Owner+"} "
	
	if imeiId == "" {
		fmt.Printf("Invalid device ID")
	}
	
	err = json.Unmarshal([]byte(json_device), &d)
	
	record, err := stub.GetState(d.IMEI)
	
	if record != nil { return nil, errors.New("Device already exists") }
	
	_, err = t.save_changes(stub, d)
	
	if err != nil { fmt.Printf("CREATEDEVICE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("imeiIds")

	if err != nil { return nil, errors.New("Unable to get imeiIds") }

	

	err = json.Unmarshal(bytes, &IMEI_Ids)

	if err != nil {	return nil, errors.New("Corrupt IMEI_Holder record") }

	IMEI_Ids.IMEIs = append(IMEI_Ids.IMEIs, imeiId)

	bytes, err = json.Marshal(IMEI_Ids)

	if err != nil { fmt.Printf("Error creating IMEI_Holder record") }

	err = stub.PutState("imeiIds", bytes)

	if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}


func (t *SimpleChainCode) createDeviceUsingForm(stub shim.ChaincodeStubInterface, args[] string) ([]byte, error) {
	
	var d Device
	var err error
	var IMEI_Ids IMEI_Holder
	var imeiId string
	
	imeiId = args[0]
	
	DeviceName  := "\"deviceName\":\""+args[1]+"\", "
	DeviceModel := "\"devicemodel\":\""+args[2]+"\", "
	DateOfManf  := "\"dateofmanf\":\"''"+args[3]+"''\" , "
	DateOfSale  := "\"dateofsale\":\"UNDEFINED\", "
	OldIMEI     := "\"oldimei\":\"UNDEFINED\", "
	IMEI_ID     := "\"imei\":\""+imeiId+"\", "
	Status     	:= "\"status\":\"CREATED\", "
	SoldBy     	:= "\"soldby\":\"UNDEFINED\", "
	Owner     	:= "\"owner\":\"VENDOR\" "
	
	json_device := " {" +DeviceName+DeviceModel+DateOfManf+DateOfSale+OldIMEI+IMEI_ID+Status+SoldBy+Owner+"} "
	
	if imeiId == "" {
		fmt.Printf("Invalid device ID")
	}
	
	err = json.Unmarshal([]byte(json_device), &d)
	
	record, err := stub.GetState(d.IMEI)
	
	if record != nil { return nil, errors.New("Device already exists") }
	
	_, err = t.save_changes(stub, d)
	
	if err != nil { fmt.Printf("CREATEDEVICE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("imeiIds")

	if err != nil { return nil, errors.New("Unable to get imeiIds") }

	

	err = json.Unmarshal(bytes, &IMEI_Ids)

	if err != nil {	return nil, errors.New("Corrupt IMEI_Holder record") }

	IMEI_Ids.IMEIs = append(IMEI_Ids.IMEIs, imeiId)

	bytes, err = json.Marshal(IMEI_Ids)

	if err != nil { fmt.Printf("Error creating IMEI_Holder record") }

	err = stub.PutState("imeiIds", bytes)

	if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}


//=================================================================================================
//  save_changes -- This function is used to save the updates of device
//=================================================================================================

func (t *SimpleChainCode) save_changes(stub shim.ChaincodeStubInterface, d Device) (bool, error) {

	bytes, err := json.Marshal(d)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting Device record: %s", err); return false, errors.New("Error converting Device record") }

	err = stub.PutState(d.IMEI, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing device record: %s", err); return false, errors.New("Error storing device record") }

	return true, nil
}

//=========================================================================================================================
//  get_device method is used to retrive device data
//=========================================================================================================================

func (t *SimpleChainCode) get_device(stub shim.ChaincodeStubInterface, imeiId string) (Device, error) {
	  var dev Device
	  bytes, err := stub.GetState(imeiId)
	  if err != nil { fmt.Printf("error while retrieving device"); return dev, errors.New("error retrieving device") }
	  err = json.Unmarshal(bytes, &dev)
	  if err != nil {fmt.Printf("failed to convert device data"); return dev, errors.New("error unmarshalling data") }
	  return dev, nil
}

func (t *SimpleChainCode) get_dev_details(stub shim.ChaincodeStubInterface, device Device) ([]byte, error){
	
	
	bytes, err := json.Marshal(device)
	
	if err != nil { return nil, errors.New("Invalid device object") }
	
//	if device.Owner  == caller {
		return bytes, nil
//	} else {
//		return nil, errors.New("Permission denied: could not return device details");
//	}
	
}

func (t *SimpleChainCode) check_unique_imei(stub shim.ChaincodeStubInterface, imei string) ([]byte, error) {
	_, err := t.get_device(stub, imei)
	if err == nil {
		return []byte("false"), errors.New("IMEI is not unique")
	} else {
		return []byte("true"), nil
	}
}

func (t *SimpleChainCode) get_devices(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytes, err := stub.GetState("imeiIds")

	if err != nil { return nil, errors.New("Unable to get imeiIds") }

	var imeiIDs IMEI_Holder

	err = json.Unmarshal(bytes, &imeiIDs)

	if err != nil {	return nil, errors.New("Corrupt IMEI_Holder") }

	result := "["

	var temp []byte
	var dev Device

	for _, imei := range imeiIDs.IMEIs {

		dev, err = t.get_device(stub, imei)

		if err != nil {return nil, errors.New("Failed to retrieve IMEI")}

		temp, err = t.get_dev_details(stub, dev)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}


func (t *SimpleChainCode) tranfer_to_WareHouse(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, consignNumber string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "VENDOR" &&
		recipientAffiliation == "WAREHOUSE" &&
		dev.Status == "CREATED"	  {
		fmt.Printf(" tranfer_to_WareHouse :: data set"); 
			dev.Status = "DELIVERED_TO_WAREHOUSE"
			dev.DateOfDelivery = time.Now().String();
			dev.ConsignmentNumber = consignNumber
	} else {
		fmt.Printf(" tranfer_to_WareHouse :: Permission denied"); 
		return nil, errors.New("error while updating device status to Delivered to warehouse"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on transfer to warehouse")}
	fmt.Printf(" tranfer_to_WareHouse :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) accept_from_vendor(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "WAREHOUSE" &&
		recipientAffiliation == "WAREHOUSE" &&
		dev.Owner == "VENDOR" && 
		dev.Status == "DELIVERED_TO_WAREHOUSE"	  {
		fmt.Printf(" accept_from_vendor"); 
			dev.Status = "Received"
			dev.Owner = "WAREHOUSE"
			dev.DateOfReceipt = time.Now().String();
	} else {
		fmt.Printf(" tranfer_to_WareHouse :: Permission denied"); 
		return nil, errors.New("error while receiving device at warehouse"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details while accept from warehouse")}
	fmt.Printf(" accept_from_vendor :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) tranfer_to_store(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, consignNumber string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "WAREHOUSE" &&
		recipientAffiliation == "STORE" &&
		dev.Status == "Received"	  {
		fmt.Printf(" tranfer_to_store :: data set"); 
			dev.Status = "DELIVERED_TO_STORE"
			dev.DateOfDelivery = time.Now().String()
			dev.ConsignmentNumber = consignNumber 
	} else {
		fmt.Printf(" tranfer_to_store :: Permission denied"); 
		return nil, errors.New("error while updating device status to Delivered to store"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on transfer to store")}
	fmt.Printf(" tranfer_to_store :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) accept_from_warehouse(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "STORE" &&
		recipientAffiliation == "STORE" &&
		dev.Owner == "WAREHOUSE" &&
		dev.Status == "DELIVERED_TO_STORE"	  {
		fmt.Printf(" accept_from_warehouse :: data set"); 
			dev.Status = "Received"
			dev.Owner = "STORE"
			dev.DateOfReceipt = time.Now().String()
			
	} else {
		fmt.Printf(" accept from warehouse :: Permission denied"); 
		return nil, errors.New("error while updating device status to received by store"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on accept from warehouse")}
	fmt.Printf(" accept_from_warehouse :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) tranfer_to_customer(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, callerName string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "STORE" &&
		recipientAffiliation == "STORE" &&
		dev.Owner == "STORE" &&
		dev.Status == "Received"	  {
		fmt.Printf(" tranfer_to_store :: data set"); 
			dev.Status = "DELIVERED_TO_CUSTOMER"
			dev.DateOfSale = time.Now().String()
			dev.SoldBy = callerName
			dev.Owner = recipientName
			
	} else {
		fmt.Printf(" tranfer_to_customer :: Permission denied"); 
		return nil, errors.New("error while updating device status to Delivered to customer"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on transfer to customer")}
	fmt.Printf(" tranfer_to_customer :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) return_from_customer(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "STORE" &&
		recipientAffiliation == "STORE" &&
		dev.Owner == "CUSTOMER" &&
		dev.Status == "DELIVERED_TO_CUSTOMER"	  {
		fmt.Printf(" tranfer_to_store :: data set"); 
			dev.Status = "RETURNED_TO_STORE"
			dev.DateOfReceipt = time.Now().String()
			dev.Owner = recipientName
	} else {
		fmt.Printf(" return_from_customer :: Permission denied"); 
		return nil, errors.New("error while updating device status to return from customer"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return from customer")}
	fmt.Printf(" return from customer :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) exchange_device(stub shim.ChaincodeStubInterface, oldDev Device, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	fmt.Printf("callerAffliation :: " + callerAffliation);
	fmt.Printf("recipientAffiliation :: " + recipientAffiliation);
	fmt.Printf("oldDev.Owner :: " + oldDev.Owner);
	fmt.Printf("oldDev.Status :: " + oldDev.Status);
	fmt.Printf("dev.Status :: " + dev.Status);
	fmt.Printf("oldDev.DeviceModel :: " + oldDev.DeviceModel);
	fmt.Printf("dev.DeviceModel :: " + dev.DeviceModel);
	fmt.Printf("model compare result :: " )
	fmt.Println(oldDev.DeviceModel == dev.DeviceModel);
	
	if  callerAffliation == "STORE" &&
		recipientAffiliation == "STORE" &&
		oldDev.Owner == "STORE" &&
		oldDev.Status == "RETURNED_TO_STORE" &&
		dev.Status == "Received" &&
		oldDev.DeviceModel == dev.DeviceModel	  {
		fmt.Printf(" exchange device :: data set"); 
			dev.Status = "Exchanged"
			dev.DateOfSale = time.Now().String()
			dev.Owner = recipientName
			dev.OldIMEI=oldDev.IMEI
	} else {
		fmt.Printf(" return_from_customer :: Permission denied"); 
		return nil, errors.New("error while updating device status to return from customer"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return from customer")}
	fmt.Printf(" return from customer :: completed"); 
	return nil, nil
}


func (t *SimpleChainCode) return_to_warehouse(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, consignNumber string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "STORE" &&
		recipientAffiliation == "WAREHOUSE" &&
		dev.Owner == "STORE" &&
		dev.Status == "RETURNED_TO_STORE"	  {
		fmt.Printf(" return_to_warehouse :: data set"); 
			dev.Status = "RETURNED_TO_WAREHOUSE"
			dev.DateOfDelivery = time.Now().String()
			dev.ConsignmentNumber = consignNumber
	} else {
		fmt.Printf(" return_to_warehouse :: Permission denied"); 
		return nil, errors.New("error while updating device status to return_to_warehouse"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return_to_warehouse")}
	fmt.Printf(" return_to_warehouse :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) return_from_store(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "WAREHOUSE" &&
		recipientAffiliation == "WAREHOUSE" &&
		dev.Owner == "STORE" &&
		dev.Status == "RETURNED_TO_WAREHOUSE"	  {
		fmt.Printf(" return_from_store :: data set"); 
			dev.Status = "Received"
			dev.DateOfReceipt = time.Now().String()
			dev.Owner = "WAREHOUSE"
	} else {
		fmt.Printf(" return_from_store :: Permission denied"); 
		return nil, errors.New("error while updating device status to return_from_store"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return_from_store")}
	fmt.Printf(" return_from_store :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) return_to_vendor(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, consignNumber string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "WAREHOUSE" &&
		recipientAffiliation == "VENDOR" &&
		dev.Owner == "WAREHOUSE" &&
		dev.Status == "Received"	  {
		fmt.Printf(" return_to_vendor :: data set"); 
			dev.Status = "RETURNED_TO_VENDOR"
			dev.DateOfDelivery = time.Now().String()
			dev.ConsignmentNumber = consignNumber
	} else {
		fmt.Printf(" return_to_vendor :: Permission denied"); 
		return nil, errors.New("error while updating device status to return from customer"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return_to_vendor")}
	fmt.Printf(" return_to_vendor :: completed"); 
	return nil, nil
}

func (t *SimpleChainCode) return_from_warehouse(stub shim.ChaincodeStubInterface, dev Device, callerAffliation string, recipientName string, recipientAffiliation string) ([]byte, error) {
	if  callerAffliation == "VENDOR" &&
		recipientAffiliation == "VENDOR" &&
		dev.Owner == "WAREHOUSE" &&
		dev.Status == "RETURNED_TO_VENDOR"	  {
		fmt.Printf(" return_from_warehouse :: data set"); 
			dev.Status = "Received"
			dev.DateOfDelivery = time.Now().String()
			dev.Owner = "VENDOR"
	} else {
		fmt.Printf(" return_from_warehouse :: Permission denied"); 
		return nil, errors.New("error while updating device status to return_from_warehouse"); 
	}
	
	_, err := t.save_changes(stub, dev)
	
	if err != nil {fmt.Printf("error while updating the status"); return nil, errors.New("error saving device details on return_from_warehouse")}
	fmt.Printf(" return_from_warehouse :: completed"); 
	return nil, nil
}

func main() {
	
	err := shim.Start(new(SimpleChainCode));
	
	if err != nil { fmt.Println("error while starting shim code"); 
	} else {
		fmt.Println("chaincode started");
	}
}
