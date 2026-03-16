using System.Net.Sockets;
using System.Text;
using System.Text.Json;

namespace WizController.Services;

public class WizService
{
    private const int WizPort = 38899;

    public async Task SetLightStateAsync(string ip, bool state)
    {
        var payload = new
        {
            method = "setState",
            @params = new { state = state }
        };

        string jsonPayload = JsonSerializer.Serialize(payload);
        byte[] data = Encoding.UTF8.GetBytes(jsonPayload);

        using var udpClient = new UdpClient();
        try
        {
            await udpClient.SendAsync(data, data.Length, ip, WizPort);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error sending to {ip}: {ex.Message}");
        }
    }

    public async Task SetBrightnessAsync(string ip, int brightness)
    {
        // Brightness range: 10-100
        var payload = new
        {
            method = "setPilot",
            @params = new { dimming = brightness }
        };

        string jsonPayload = JsonSerializer.Serialize(payload);
        byte[] data = Encoding.UTF8.GetBytes(jsonPayload);

        using var udpClient = new UdpClient();
        try
        {
            await udpClient.SendAsync(data, data.Length, ip, WizPort);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error sending to {ip}: {ex.Message}");
        }
    }
}
