package de.unibonn.mss15.sensorlogger;

import android.util.Log;

import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken;
import org.eclipse.paho.client.mqttv3.MqttCallback;
import org.eclipse.paho.client.mqttv3.MqttClient;
import org.eclipse.paho.client.mqttv3.MqttConnectOptions;
import org.eclipse.paho.client.mqttv3.MqttException;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence;
import com.google.gson.Gson;

/**
 * Created by Farshid on 05-Sep-15.
 */
public class MQTTClient implements MqttCallback {

    MemoryPersistence memPer;
    MqttClient client;
    final String BrokerURI = "tcp://192.168.1.42:1883";
    //final String BrokerURI = "tcp://iot.eclipse.org:1883";

    public MQTTClient(){
        connect();
    }

    private void connect(){
        memPer = new MemoryPersistence();
        try
        {
            client = new MqttClient(BrokerURI, MqttClient.generateClientId(), null);
            client.setCallback(this);
        }
        catch (MqttException e1)
        {
            e1.printStackTrace();
        }

        MqttConnectOptions options = new MqttConnectOptions();
        try
        {
            client.connect(options);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Connection attempt failed with reason code = " + e.getReasonCode() + ":" + e.getCause());
        }
    }

    public void publish(String msg){
        try
        {
            MqttMessage message = new MqttMessage();
            message.setQos(1);
            message.setPayload(msg.getBytes());
            client.publish("sensors/data", message);
        }
        catch (MqttException e)
        {
            Log.d(getClass().getCanonicalName(), "Publish failed with reason code = " + e.getReasonCode());
        }
    }

    public void publicJSON(long t, int e, String n, int axes, float... values){
        try{
            Entry entry = new Entry("?",t,e,n,axes,values);
            Gson gson = new Gson();
            String json = gson.toJson(entry);
            publish(json);
        }
        catch(IllegalArgumentException|SecurityException|IllegalAccessException|NoSuchFieldException ex){
            Log.v("Exception", ex.getMessage());
        }
    }


    @Override
    public void connectionLost(Throwable cause)
    {
        Log.d("MQTT", "MQTT Server connection lost" + cause.getMessage());
    }
    @Override
    public void messageArrived(String topic, MqttMessage message)
    {
        Log.d("MQTT", "Message arrived:" + topic + ":" + message.toString());
    }
    @Override
    public void deliveryComplete(IMqttDeliveryToken token)
    {
        Log.d("MQTT", "Delivery complete");
    }
}
