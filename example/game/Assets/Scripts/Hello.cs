using TMPro;
using UnityEngine;

public class Hello : MonoBehaviour
{
    public TMP_Text text;

    // Start is called once before the first execution of Update after the MonoBehaviour is created
    void Start()
    {
        text.text = "Hello, Dirk!";
    }
}
